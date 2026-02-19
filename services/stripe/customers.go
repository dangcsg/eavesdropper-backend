package stripe

import (
	"context"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/customer"
)

// Initialize Stripe - call this during your application startup
func InitStripe(apiKey string) {
	stripe.Key = apiKey
}

// CreateCustomer creates a new customer in Stripe
func CreateCustomer(ctx context.Context, email, name string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
	}

	return customer.New(params)
}
