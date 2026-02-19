package requests

type CheckoutSession struct {
	SuccessURL string
	CancelURL  string
	PriceID    string
	CustomerID string
}

type PortalSession struct {
	UserID     string
	CustomerID string
	ReturnUrl  string
}
