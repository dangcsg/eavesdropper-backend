package api

import (
	"eavesdropper/api/handlers"
	m "eavesdropper/api/middlewares"
	"log"
	"net/http"
)

type Router struct {
	mux *http.ServeMux
}

// address is in the format ":8080"
func (r *Router) Start(address string) {
	r.mux = http.NewServeMux()

	r.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })

	r.mux.HandleFunc("POST /users", m.ValidateToken(handlers.AddUser))
	r.mux.HandleFunc("PUT /users/{id}", m.ValidateOwnership(handlers.UpdateUser))
	r.mux.HandleFunc("GET /users/{id}", m.ValidateOwnership(handlers.GetUser))
	r.mux.HandleFunc("GET /users/{id}/billing-cycle-usage", m.ValidateOwnership(handlers.GetUserBillingCycleUsage))
	r.mux.HandleFunc("GET /users/handle/availability", m.ValidateToken(handlers.HandleAvailable))

	r.mux.HandleFunc("GET /users/{id}/transcripts", m.ValidateOwnership(handlers.GetUserTranscripts))
	r.mux.HandleFunc("GET /users/{id}/transcripts/{tId}", m.ValidateOwnership(handlers.GetUserTranscript))
	r.mux.HandleFunc("POST /users/{id}/transcript/{tId}/whitelist", m.ValidateOwnership(handlers.AddUsersToTranscriptWhitelist))
	r.mux.HandleFunc("PUT /users/{id}/transcripts/{tId}/visibility", m.ValidateOwnership(handlers.UpdateTranscriptVisibility))
	r.mux.HandleFunc("PUT /users/{id}/transcripts/{tId}/tittle", m.ValidateOwnership(handlers.UpdateTranscriptTitle))
	r.mux.HandleFunc("DELETE /users/{id}/transcript/{tId}/whitelist", m.ValidateOwnership(handlers.RemoveUsersFromTranscriptWhitelist))
	r.mux.HandleFunc("GET /users/{id}/transcript/{tId}/whitelist", m.ValidateOwnership(handlers.GetTranscriptWhitelist))

	r.mux.HandleFunc("GET /subscription-plans", m.ValidateToken(handlers.GetSubscriptionPlans))

	r.mux.HandleFunc("POST /stripe/webhook", handlers.StripeWebhookHandler)
	r.mux.HandleFunc("POST /stripe/customer", m.ValidateToken(handlers.CreateStripeCustomer))
	r.mux.HandleFunc("POST /stripe/session/checkout", m.ValidateToken(handlers.CreateCheckoutSession))
	r.mux.HandleFunc("POST /stripe/session/portal", m.ValidateToken(handlers.CreatePortalSession))

	r.mux.HandleFunc("POST /transcripts", m.ValidateToken(handlers.Transcribe))
	r.mux.HandleFunc("GET /transcripts/{id}", handlers.GetTranscript)

	// Sets the global middlewares
	handler := m.CorsMiddleware(m.LoggingMiddleware(r.mux))

	log.Printf("listening on %s", address)
	log.Fatal(http.ListenAndServe(address, handler))
}
