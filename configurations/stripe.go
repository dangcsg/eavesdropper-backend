package configurations

import "os"

func GetStripeKey() string {
	if SelectedBackendMode == Production {
		return os.Getenv("STRIPE_SECRET_KEY_PROD")
	}
	return os.Getenv("STRIPE_SECRET_KEY_DEV")
}

// The secret for the webhook handler
func StripeEndpointSecret() string {
	if SelectedBackendMode == Production {
		return os.Getenv("STRIPE_WEBHOOK_SECRET_PROD")
	}
	return os.Getenv("STRIPE_WEBHOOK_SECRET_DEV")
}
