package subscriptions

import (
	cnfgs "eavesdropper/configurations"
	"eavesdropper/dtos/responses"
)

func GetSubscriptionPlans() ([]responses.SubscriptionPlan, error) {
	plans := []responses.SubscriptionPlan{}

	for i, tier := range cnfgs.SubscriptionTiersList() {
		priceID, err := cnfgs.GetPriceID(tier)
		if err != nil {
			return nil, err
		}

		planName, err := cnfgs.GetPlanName(tier)
		if err != nil {
			return nil, err
		}

		monthlyLimit, err := cnfgs.GetMonthlyAudioSeconds(tier)
		if err != nil {
			return nil, err
		}

		var freeAudioSeconds = 0
		if tier == cnfgs.FreeTrial {
			freeAudioSeconds = cnfgs.FreeAudioSeconds
		}

		plan := responses.SubscriptionPlan{
			ID:                          i,
			PriceID:                     priceID,
			Name:                        planName,
			TranscriptionMonthlySeconds: monthlyLimit,
			FreeTranscriptionSeconds:    freeAudioSeconds,
		}
		plans = append(plans, plan)
	}

	return plans, nil
}
