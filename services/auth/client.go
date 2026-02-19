package auth

import (
	"context"
	"eavesdropper/services/firebase"
	"log"

	"firebase.google.com/go/v4/auth"
)

var Auth = newAuthInstance()

type AuthenticationApp struct {
	Ctx    context.Context
	Client *auth.Client
}

func newAuthInstance() *AuthenticationApp {

	client, err := firebase.FirebaseInstance.Client.Auth(firebase.FirebaseInstance.Ctx)
	if err != nil {
		log.Fatal("failed to initialize firebase authentication client: ", err)
	}

	authenticator := new(AuthenticationApp)
	authenticator.Client = client
	authenticator.Ctx = firebase.FirebaseInstance.Ctx

	return authenticator
}
