package handlers

import (
	apiErr "eavesdropper/api/error"
	"eavesdropper/api/middlewares"
	cnfgs "eavesdropper/configurations"
	"eavesdropper/dtos/requests"
	"eavesdropper/dtos/responses"
	"eavesdropper/services/auth"
	stripeService "eavesdropper/services/stripe"
	"eavesdropper/services/users"
	"eavesdropper/services/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/stripe/stripe-go/v82"
	billingSession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

// todo move business logic to service layer.

func CreateStripeCustomer(w http.ResponseWriter, r *http.Request) {

	token, ok := r.Context().Value(middlewares.AuthTokenKey).(string)
	if !ok {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to find auth token request context.")
		return
	}
	userID, err := auth.GetUserID(r.Context(), token)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusUnauthorized, "", "Failed to get userID from token: "+err.Error())
		return
	}

	customerID, err := users.AddStripe(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to add stripe to user %s", err.Error())
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to create stripe customer"+err.Error())
		return
	}

	response := map[string]interface{}{
		"customerID": customerID,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// CreateCheckoutSession creates a Stripe Checkout Session
func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req requests.CheckoutSession
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Failed to deserialize request"+err.Error())
		return
	}

	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(req.SuccessURL),
		CancelURL:  stripe.String(req.CancelURL),
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(req.PriceID),
				Quantity: stripe.Int64(1),
			},
		},
		Customer: stripe.String(req.CustomerID),
	}

	s, err := session.New(params)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to create stripe session: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses.CheckoutSession{
		ID:           s.ID,
		ClientSecret: s.ClientSecret,
	})
}

// CreatePortalSession creates a Stripe Billing Portal Session
func CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req requests.PortalSession
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Failed to unmarshal portal session object "+err.Error())
		return
	}

	fmt.Println("creating portal session from req:")
	utils.PrintJSON(req)

	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(req.CustomerID),
		ReturnURL: stripe.String(req.ReturnUrl),
		// Metadata can be added here if you decide to use it:
		// Metadata: map[string]string{"userId": req.UserId},
	}

	ps, err := billingSession.New(params)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to create billing session: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(responses.PortalSession{
		URL: ps.URL,
	})
}

// StripeWebhookHandler handles incoming Stripe webhooks
func StripeWebhookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Println("Got stripe event")

	// Read raw payload
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "failed to read request body: "+err.Error())
		log.Println("Stripe error: failed to read request body")
		return
	}

	// Verify signature
	signature := r.Header.Get("Stripe-Signature")
	if signature == "" {
		log.Println("Stripe error: missing stripe signature")
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Missing the stripe signature")
		return
	}

	event, err := stripeService.WebhookHandler(payload, signature, cnfgs.StripeEndpointSecret())
	if err != nil {
		log.Println("invalid stripe signature: " + err.Error())
		apiErr.WriteJSONError(w, http.StatusBadRequest, "", "invalid stripe signature: "+err.Error())
		return
	}

	log.Printf("got stripe event type: %s", event.Type)

	switch event.Type {

	case "invoice.paid":
		log.Println("On invoice.paid webhook event")

		var invoice stripe.Invoice
		if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
			fmt.Println("failed to unmarshal: ", err.Error())
			apiErr.WriteJSONError(w, http.StatusBadRequest, "", "Failed to unmarshal stripe invoice")
			return
		}

		if jsonData, _ := json.MarshalIndent(invoice, "", "  "); len(jsonData) > 0 {
			log.Println("invoice json:")
			log.Println(string(jsonData))
		}

		if len(invoice.Lines.Data) == 0 {
			apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get the priceID from the invoice data")
			return
		}
		priceID := invoice.Lines.Data[0].Pricing.PriceDetails.Price

		hasSubID := (invoice.Parent != nil &&
			invoice.Parent.SubscriptionDetails != nil &&
			invoice.Parent.SubscriptionDetails.Subscription != nil)

		if hasSubID {
			log.Println("Got non nil subscription: ", invoice.Parent.SubscriptionDetails.Subscription.ID)

			customerID := invoice.Customer.ID
			subscriptionID := invoice.Parent.SubscriptionDetails.Subscription.ID

			subsToKeep := []string{subscriptionID}
			go stripeService.CancelUserSubsctiptions(customerID, subsToKeep)

			subscriptionTier, err := cnfgs.GetSubscriptionTier(priceID)
			if err != nil {
				apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get the subscription tier associated with the priceID: "+err.Error())
				return
			}

			err = users.AddSubscriptionToUser(customerID, subscriptionID, int(subscriptionTier))
			if err != nil {
				log.Println("Failed to save subscription in bd: ", err.Error())
				apiErr.WriteJSONError(w, http.StatusInternalServerError, "", "Failed to get to add sub to user: "+err.Error())
				return
			}

			if invoice.Attempted {
				log.Printf("Payment recovered for subscription: Customer ID: %s, Subscription ID: %s\n",
					customerID, subscriptionID)
			}
		}

	case "invoice.payment_failed":
		log.Println("on invoice payment failed webhook")
		// Optionally handle failed payment details here

	case "checkout.session.completed":
		log.Println("on checkout session completed webhook")

	case "customer.subscription.created":
		log.Println("on subscription created webhook")

	case "customer.subscription.updated":
		log.Println("on subscription updated webhook")

	case "customer.subscription.deleted":
		log.Println("on subscription deleted webhook")

	case "invoice.payment_succeeded":
		log.Println("on invoice payment succeded webhook")
	}

	// Send the acknowledgment once (after processing)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]bool{"received": true})
}
