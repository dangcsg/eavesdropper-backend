package transcribe

import (
	"context"
	cnfgs "eavesdropper/configurations"
	"eavesdropper/errs"
	db "eavesdropper/services/data/firestore/operations"
	"eavesdropper/services/stripe"
	"eavesdropper/services/users"
	"errors"
	"fmt"
	"time"

	s "github.com/stripe/stripe-go/v82"
	"google.golang.org/genai"
)

func TranscriptionAllowed(
	ctx context.Context,
	userID string,
	audioSeconds int,
) (
	subscriptionID string,
	consumedFreeSeconds int,
	err error,
) {

	user, err := db.GetUser(userID)
	if err != nil {
		return "", 0, err
	}

	if user.FreeTranscriptionSeconds >= audioSeconds {
		return "", audioSeconds, nil
	}

	// Checks if the user has a valid stripe subscription
	if user.StripeCustomerID == "" {
		return "", 0, errs.ErrUserHasNoStripeAccount
	}
	timeb4sub := time.Now()
	stripeSubscription, err := stripe.CheckUserSubscription(ctx, user.StripeCustomerID)
	if err != nil {
		return "", 0, err
	}
	fmt.Println("took ", time.Since(timeb4sub), " to check user subscription")
	if stripeSubscription == nil {
		return "", 0, errs.ErrUserHasNoActiveSubscription
	}
	if stripeSubscription.Status == s.SubscriptionStatusPastDue {
		return "", 0, errs.ErrUserSubscriptionIsExpired
	}
	if stripeSubscription.Status != s.SubscriptionStatusActive {
		return "", 0, errs.ErrUserHasNoActiveSubscription
	}

	subscription, err := db.GetSubscription(userID, stripeSubscription.ID)
	if err != nil {
		return "", 0, err
	}
	if subscription == nil {
		return "", 0, errors.New("Subscription not found")
	}
	subscriptionTier, err := cnfgs.GetSubscriptionTier(stripeSubscription.PriceID)
	if err != nil {
		return "", 0, err
	}
	transcriptionSecondsLimit, err := cnfgs.GetMonthlyAudioSeconds(subscriptionTier)
	if err != nil {
		return "", 0, err
	}

	billingCycleUsage, err := users.GetCurrentBillingCycleUsage(ctx, userID)
	if err != nil {
		return "", 0, err
	}

	audioSecondsToPay := audioSeconds - user.FreeTranscriptionSeconds
	if billingCycleUsage.ConsumedInputAudioSeconds+audioSecondsToPay > transcriptionSecondsLimit {
		return subscription.ID, 0, errs.ErrExceededSubscriptionTranscriptionLimits
	}

	return subscription.ID, user.FreeTranscriptionSeconds, nil
}

func TranscribeAudioFile(ctx context.Context, audioFilePath, audioFormat string) (*genai.GenerateContentResponse, error) {

	fmt.Printf("transcribing audio with path %s and format %s", audioFilePath, audioFormat)

	client, err := getGenaiClient(ctx)
	if err != nil {
		return nil, err
	}

	uploadedFile, err := client.Files.UploadFromPath(ctx, audioFilePath, nil)
	if err != nil {
		return nil, err
	}

	err = awaitActiveFile(ctx, client, uploadedFile.Name)
	if err != nil {
		fmt.Println("got err await ready file: ", err)
	}

	parts := []*genai.Part{
		genai.NewPartFromText(`
			Transcribe the audio. Include speaker labels such as:
			Speaker 1: ...
			Speaker 2: ...
			Try to distinguish between speakers whenever the voice changes.
			`),
		genai.NewPartFromURI(uploadedFile.URI, audioFormat),
	}
	contents := []*genai.Content{
		genai.NewContentFromParts(parts, genai.RoleUser),
	}

	return client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		contents,
		nil,
	)
}

func awaitActiveFile(ctx context.Context, client *genai.Client, fileName string) error {
	fmt.Printf("awaiting for file %s to be ready", fileName)

	var err error = nil
	var file *genai.File = nil
	for i := 0; i < 10; i++ {

		time.Sleep(300 * time.Millisecond)

		file, err = client.Files.Get(ctx, fileName, nil)
		if err != nil {
			return fmt.Errorf("error checking file status: %w", err)
		}

		if file.State == genai.FileStateActive {
			return nil
		}
	}

	if err != nil {
		return err
	}
	if file.State == genai.FileStateFailed {
		return fmt.Errorf("Failed to upload file! Got error %v", file.Error)
	}

	return errors.New("Timeout ocurred while awaiting the ready file state")

}
