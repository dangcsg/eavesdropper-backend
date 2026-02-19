package handlers

import (
	apiErr "eavesdropper/api/error"
	"eavesdropper/services/subscriptions"
	"encoding/json"
	"net/http"
)

func GetSubscriptionPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := subscriptions.GetSubscriptionPlans()
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get subscription plans: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}