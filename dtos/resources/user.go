package resources

import "time"

type User struct {
	ID                       string
	Handle                   string
	Email                    string
	FirstName                string
	LastName                 string
	StripeCustomerID         string
	PrefersDarkMode          bool
	FreeTranscriptionSeconds int
	CreatedAt                time.Time
}
