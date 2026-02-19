package stripe

import (
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// WebhookHandler processes Stripe webhook events
func WebhookHandler(payload []byte, signature string, endpointSecret string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, endpointSecret)
}
