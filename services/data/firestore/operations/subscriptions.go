package operations

import (
	"context"
	"eavesdropper/dtos/resources"
	db "eavesdropper/services/data/firestore/client"
	"eavesdropper/services/data/firestore/collections"
	"errors"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/firestore"
)

func SaveStripeSubscription(userID, subscriptionID, customerID string, planID int) error {

	subscriptionRef := collections.Subscriptions(userID).Doc(subscriptionID)

	transaction := func(ctx context.Context, tx *firestore.Transaction) error {

		snap, err := tx.Get(subscriptionRef)
		if err != nil && status.Code(err) != codes.NotFound {
			return err
		}
		if snap.Exists() {
			return errors.New("Subscription already created")
		}

		subscription := &resources.Subscription{
			ID:                   uuid.NewString(),
			UserID:               userID,
			StripeSubscriptionID: subscriptionID,
			StripeCustomerID:     customerID,
			PlanID:               planID, // todo: maybe get rid of this, only ref sub? See if I have the plan info when calling this
		}

		return tx.Create(subscriptionRef, subscription)
	}

	return db.FirestoreInstance.Client.RunTransaction(context.Background(), transaction)
}

// Deprecated. Now it's impicit in the transcripts
// func AccountUsage(userID, subscriptionID string, usage dtos.TranscriptUsage) error {

// 	ctx := context.Background()

// 	snap, err := collections.Subscriptions(userID).Doc(subscriptionID).Get(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	subscription := new(resources.Subscription)
// 	err = snap.DataTo(subscription)
// 	if err != nil {
// 		return err
// 	}

// 	subscription.TranscriptionsCount += 1
// 	subscription.ConsumedInputAudioSeconds += usage.ConsumedInputAudioSeconds
// 	subscription.ConsumedAudioTokens += usage.ConsumedAudioTokens
// 	subscription.ConsumedTotalInputTokens += usage.ConsumedTotalInputTokens
// 	subscription.ConsumedOutputTokens += usage.ConsumedOutputTokens

// 	_, err = collections.Subscriptions(userID).Doc(subscriptionID).Set(ctx, subscription)

// 	return err
// }

func GetSubscription(userID, subscriptionID string) (*resources.Subscription, error) {

	snap, err := collections.Subscriptions(userID).Doc(subscriptionID).Get(context.Background())
	if err != nil {
		return nil, err
	}

	sub := new(resources.Subscription)
	err = snap.DataTo(sub)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
