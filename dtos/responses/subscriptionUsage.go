package responses

import "time"

type SubscriptionUsage struct {
	SubscriptonPlanName           string
	RenewsAt                      *time.Time `json:"RenewsAt,omitempty"`
	SubscriptionMonthlyMinutes    int
	TranscriptsCount              int
	ConsumedInputAudioSeconds     int
	ConsumedFreeInputAudioSeconds int
	ConsumedPaidInputAudioSeconds int
	ConsumedAudioTokens           int
	ConsumedTotalInputTokens      int
	ConsumedOutputTokens          int
}
