package firebase

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
)

var FirebaseInstance = NewFirebaseInstance()

type Firebase struct {
	Ctx    context.Context
	Client *firebase.App
}

func NewFirebaseInstance() *Firebase {
	ctx := context.Background()

	// Initializes the client
	instance, err := firebase.NewApp(ctx, nil)
	if err != nil {
		log.Fatal("failed to open firebaseApp on development mode: ", err)
	}
	var firebaseApp = instance

	// Builds the firebaseClientInstance struct
	firebaseClientInstance := new(Firebase)
	firebaseClientInstance.Ctx = ctx
	firebaseClientInstance.Client = firebaseApp

	return firebaseClientInstance
}
