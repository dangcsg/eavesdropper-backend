package users

import (
	"context"
	cnfgs "eavesdropper/configurations"
	"eavesdropper/dtos/requests"
	"eavesdropper/dtos/resources"
	"eavesdropper/dtos/responses"
	firestoreClient "eavesdropper/services/data/firestore/client"
	"eavesdropper/services/data/firestore/collections"
	"eavesdropper/services/data/firestore/operations"
	db "eavesdropper/services/data/firestore/operations"
	"eavesdropper/services/stripe"
	"fmt"
	"math"
	"time"

	"cloud.google.com/go/firestore"
)

func AddUser(id string, data *requests.NewUser) error {

	user := &resources.User{
		ID:                       id,
		Email:                    data.Email,
		PrefersDarkMode:          data.PrefersDarkMode,
		FreeTranscriptionSeconds: cnfgs.FreeAudioSeconds,
		StripeCustomerID:         "",
		CreatedAt:                time.Now(),
	}

	return db.CreateUser(user)
}

func AddStripe(ctx context.Context, userID string) (customerID string, err error) {
	user, err := db.GetUser(userID)
	if err != nil {
		return "", err
	}

	customer, err := stripe.CreateCustomer(
		ctx,
		user.Email,
		fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	)
	if err != nil {
		return "", err
	}

	user.StripeCustomerID = customer.ID

	err = db.UpdateUser(user)
	if err != nil {
		return "", err
	}

	return customer.ID, nil
}

func AddSubscriptionToUser(customerID, subscriptionID string, planID int) error {

	fmt.Println("Adding sub to user with customer id: ", customerID)

	user, err := db.GetUserWithStripe(customerID)
	if err != nil {
		return err
	}

	fmt.Println("got user: ", user.ID)

	fmt.Println("Saving sub id: ", subscriptionID)
	err = operations.SaveStripeSubscription(user.ID, subscriptionID, user.StripeCustomerID, planID)
	if err != nil {
		return err
	}

	return nil
}

func DecrementFreeTier(userID string, consumedSeconds int) error {

	user, err := GetUser(userID)
	if err != nil {
		return err
	}

	user.FreeTranscriptionSeconds = int(math.Max(float64(user.FreeTranscriptionSeconds-consumedSeconds), 0))

	return db.UpdateUser(user)
}

func GetUser(id string) (*resources.User, error) {
	return db.GetUser(id)
}

func HandleIsAvailable(ctx context.Context, handle string) (bool, error) {
	return db.HandleIsAvailable(ctx, handle)
}

func UpdateUser(id string, data *requests.UpdateUser) error {
	userRef := collections.Users.Doc(id)

	transaction := func(ctx context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(userRef)
		if err != nil {
			return fmt.Errorf("failed to get user: %v", err)
		}

		user := new(resources.User)
		if err := snap.DataTo(user); err != nil {
			return fmt.Errorf("failed to parse user data: %v", err)
		}

		if data.Handle != nil && user.Handle == "" {
			user.Handle = *data.Handle
		}

		if data.FirstName != nil {
			user.FirstName = *data.FirstName
		}
		if data.LastName != nil {
			user.LastName = *data.LastName
		}
		if data.PrefersDarkMode != nil {
			user.PrefersDarkMode = *data.PrefersDarkMode
		}

		return tx.Set(userRef, user)
	}

	return firestoreClient.FirestoreInstance.Client.RunTransaction(context.Background(), transaction)
}

func GetUserStripeSub(id string) (*stripe.SubscriptionStatus, error) {

	ctx := context.Background()

	user, err := db.GetUser(id)
	if err != nil {
		return nil, err
	}

	if user.StripeCustomerID == "" {
		return nil, err
	}

	return stripe.CheckUserSubscription(ctx, user.StripeCustomerID)
}

func GetCurrentBillingCycleUsage(ctx context.Context, userID string) (*responses.SubscriptionUsage, error) {

	stripeSub, err := GetUserStripeSub(userID)
	if err != nil {
		return nil, err
	}

	var planName = "Free Tier"
	var billingCycleStart = time.Unix(0, 0).UTC()             // Epoch
	var billingCycleEnd = time.Now().Add(30 * 24 * time.Hour) // Any time in the future
	var subscriptionMonthlyLimit = cnfgs.FreeAudioSeconds
	var renewalDate *time.Time
	if stripeSub != nil {
		subscriptionTier, err := cnfgs.GetSubscriptionTier(stripeSub.PriceID)
		if err != nil {
			return nil, err
		}
		planName, err = cnfgs.GetPlanName(subscriptionTier)
		if err != nil {
			return nil, err
		}
		subscriptionMonthlyLimit, err = cnfgs.GetMonthlyAudioSeconds(subscriptionTier)
		if err != nil {
			return nil, err
		}
		billingCycleStart = stripeSub.BillingCycleStart
		billingCycleEnd = stripeSub.BillingCycleEnd
		renewalDate = &stripeSub.BillingCycleEnd
	}

	fmt.Printf("\nGetting transcripts for user %s from %v, to %v", userID, billingCycleStart, billingCycleEnd)
	transcripts, err := db.GetTranscripts(ctx, userID, billingCycleStart, billingCycleEnd)
	if err != nil {
		return nil, err
	}

	fmt.Printf("got %v\n", transcripts)

	usage := &responses.SubscriptionUsage{
		SubscriptonPlanName:        planName,
		TranscriptsCount:           len(transcripts),
		SubscriptionMonthlyMinutes: subscriptionMonthlyLimit,
		RenewsAt:                   renewalDate,
	}

	fmt.Printf("Calculating usage for %d transcripts\n", usage.TranscriptsCount)

	for _, t := range transcripts {
		usage.ConsumedInputAudioSeconds += t.ConsumedInputAudioSeconds
		usage.ConsumedFreeInputAudioSeconds += t.ConsumedFreeAudioSeconds
		usage.ConsumedPaidInputAudioSeconds += (t.ConsumedInputAudioSeconds - t.ConsumedFreeAudioSeconds)
		usage.ConsumedTotalInputTokens += t.ConsumedInputTokens
		usage.ConsumedOutputTokens += t.ConsumedOutputTokens
	}

	return usage, nil
}
