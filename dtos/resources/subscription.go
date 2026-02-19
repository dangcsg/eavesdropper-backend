package resources

type Subscription struct {
	ID                   string
	UserID               string
	PlanID               int    // todo switch to enum
	StripeCustomerID     string // unecessary?
	StripeSubscriptionID string
}
