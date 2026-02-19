package configurations

import (
	"fmt"
)

const FreeAudioSeconds = 30 * 60

type SubscriptionTier int

const (
	FreeTrial SubscriptionTier = iota
	StarterTier
	ProTier
	PremiumTier
)

// PlanConfig holds all configuration for a subscription tier
type PlanConfig struct {
	Name                string
	PriceIDTest         string
	PriceIDProd         string
	MonthlyAudioSeconds int
}

// planConfigs is the single source of truth for all plan configurations
var planConfigs = map[SubscriptionTier]PlanConfig{
	FreeTrial: {
		Name:                "Free Trial",
		PriceIDTest:         "",
		PriceIDProd:         "",
		MonthlyAudioSeconds: 0,
	},
	StarterTier: {
		Name:                "Starter Plan",
		PriceIDTest:         "price_1SEE8qDBXz9Kq4HnJnQgI3rR",
		PriceIDProd:         "3",
		MonthlyAudioSeconds: 5 * 60 * 60, // 5 hours
	},
	ProTier: {
		Name:                "Pro Plan",
		PriceIDTest:         "price_1SEHXaDBXz9Kq4HnL8wPPi9z",
		PriceIDProd:         "1",
		MonthlyAudioSeconds: 60 * 60, // 1 hour
	},
	PremiumTier: {
		Name:                "Premium Plan",
		PriceIDTest:         "price_1SEHZrDBXz9Kq4HnqqsH18xl",
		PriceIDProd:         "2",
		MonthlyAudioSeconds: 60 * 60, // 1 hour
	},
}

// Reverse lookup maps, built once at initialization
var (
	priceToTierTest map[string]SubscriptionTier
	priceToTierProd map[string]SubscriptionTier
)

func init() {
	priceToTierTest = make(map[string]SubscriptionTier)
	priceToTierProd = make(map[string]SubscriptionTier)

	for tier, config := range planConfigs {
		if config.PriceIDTest != "" {
			priceToTierTest[config.PriceIDTest] = tier
		}
		if config.PriceIDProd != "" {
			priceToTierProd[config.PriceIDProd] = tier
		}
	}
}

// SubscriptionTiersList returns all available tiers
func SubscriptionTiersList() []SubscriptionTier {
	return []SubscriptionTier{FreeTrial, StarterTier, ProTier, PremiumTier}
}

// GetPriceID returns the Stripe price ID for a given tier and environment
func GetPriceID(tier SubscriptionTier) (string, error) {
	config, exists := planConfigs[tier]
	if !exists {
		return "", fmt.Errorf("unknown subscription tier: %d", tier)
	}

	switch SelectedBackendMode {
	case Development:
		return config.PriceIDTest, nil
	case Production:
		return config.PriceIDProd, nil
	default:
		return "", fmt.Errorf("unknown environment: %d", SelectedBackendMode)
	}
}

// GetSubscriptionTier returns the tier for a given price ID and environment
func GetSubscriptionTier(priceID string) (SubscriptionTier, error) {
	var lookupMap map[string]SubscriptionTier

	switch SelectedBackendMode {
	case Development:
		lookupMap = priceToTierTest
	case Production:
		lookupMap = priceToTierProd
	default:
		return FreeTrial, fmt.Errorf("unknown environment: %d", SelectedBackendMode)
	}

	tier, exists := lookupMap[priceID]
	if !exists {
		return FreeTrial, fmt.Errorf("unknown price ID: %s", priceID)
	}

	return tier, nil
}

// GetPlanName returns the display name for a subscription tier
func GetPlanName(tier SubscriptionTier) (string, error) {
	config, exists := planConfigs[tier]
	if !exists {
		return "", fmt.Errorf("unknown subscription tier: %d", tier)
	}
	return config.Name, nil
}

// GetPlanNameByPriceID returns the plan name for a given price ID
func GetPlanNameByPriceID(priceID string) (string, error) {
	tier, err := GetSubscriptionTier(priceID)
	if err != nil {
		return "", err
	}
	return GetPlanName(tier)
}

// GetMonthlyAudioSeconds returns the monthly audio limit for a tier
func GetMonthlyAudioSeconds(tier SubscriptionTier) (int, error) {
	config, exists := planConfigs[tier]
	if !exists {
		return 0, fmt.Errorf("unknown subscription tier: %d", tier)
	}
	return config.MonthlyAudioSeconds, nil
}

// GetMonthlyAudioSecondsByPriceID returns the monthly limit for a price ID
func GetMonthlyAudioSecondsByPriceID(priceID string) (int, error) {
	tier, err := GetSubscriptionTier(priceID)
	if err != nil {
		return 0, err
	}
	return GetMonthlyAudioSeconds(tier)
}

// String implements Stringer for better debugging
func (s SubscriptionTier) String() string {
	if config, exists := planConfigs[s]; exists {
		return config.Name
	}
	return fmt.Sprintf("UnknownTier(%d)", s)
}
