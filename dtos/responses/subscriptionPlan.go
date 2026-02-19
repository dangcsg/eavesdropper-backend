package responses

type SubscriptionPlan struct {
	ID                          int
	PriceID                     string
	Name                        string
	TranscriptionMonthlySeconds int
	FreeTranscriptionSeconds    int
}
