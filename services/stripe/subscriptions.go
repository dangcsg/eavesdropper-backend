package stripe

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/subscription"
)

type SubscriptionStatus struct {
	ID      string
	Status  stripe.SubscriptionStatus
	PriceID string
	BillingCycleStart,
	BillingCycleEnd time.Time
}

// Retrieves the current active subscription for a user
// Returns nil, nil if none is found.
func CheckUserSubscription(ctx context.Context, stripeCustomerID string) (*SubscriptionStatus, error) {

	iter := subscription.List(&stripe.SubscriptionListParams{
		Customer: stripe.String(stripeCustomerID),
		Status:   stripe.String("active"),
	})

	for iter.Next() {
		sub := iter.Subscription()

		if sub.Status != stripe.SubscriptionStatusActive {
			continue
		}

		if len(sub.Items.Data) == 0 {
			continue
		}

		// Get the first (and only) active subscription item
		item := sub.Items.Data[0]

		return &SubscriptionStatus{
			ID:                sub.ID,
			Status:            sub.Status,
			PriceID:           item.Price.ID,
			BillingCycleStart: time.Unix(item.CurrentPeriodStart, 0),
			BillingCycleEnd:   time.Unix(item.CurrentPeriodEnd, 0),
		}, nil
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("error checking subscriptions: %w", err)
	}

	return nil, nil
}

func CancelUserSubsctiptions(customerID string, subsToIgnore []string) []error {

	// Create params to list active subscriptions for this customer
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
		Status:   stripe.String("active"),
	}

	// Create a map from the ignore list for O(1) lookups
	ignoreMap := make(map[string]bool)
	for _, subID := range subsToIgnore {
		ignoreMap[subID] = true
	}

	// Get subscription iterator
	iter := subscription.List(params)

	// Track if we encounter any errors
	errors := []error{}

	// Iterate through all active subscriptions
	for iter.Next() {
		sub := iter.Subscription()

		if ignoreMap[sub.ID] {
			continue
		}

		_, err := subscription.Cancel(sub.ID, &stripe.SubscriptionCancelParams{})
		if err != nil {
			errors = append(errors, err)
			log.Printf("Error canceling subscription %s: %v", sub.ID, err)
		}
	}

	// Check for errors in the iterator itself
	if err := iter.Err(); err != nil {
		errors = append(errors, err)
	}

	return errors
}
