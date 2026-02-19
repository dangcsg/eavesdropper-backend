package main

import (
	"eavesdropper/api"
	config "eavesdropper/configurations"
	"eavesdropper/services/stripe"
	"log"
	"os"
)

func main() {

	stripe.InitStripe(config.GetStripeKey())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // local default
	}
	addr := ":" + port

	log.Printf("Starting server on %s ...", addr)
	router := api.Router{}
	router.Start(addr)
}
