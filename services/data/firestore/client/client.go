package firestore

import (
	"context"
	firebase "eavesdropper/services/firebase"
	"log"

	"cloud.google.com/go/firestore"
)

var FirestoreInstance = NewFirestoreInstance()

type FirestoreApp struct {
	Ctx    context.Context
	Client *firestore.Client
}

func NewFirestoreInstance() *FirestoreApp {

	// Gets the firestore client
	firestoreClient, err := firebase.FirebaseInstance.Client.Firestore(firebase.FirebaseInstance.Ctx)
	if err != nil {
		log.Fatal("failed to open firestore client: ", err)
	}

	// Builds the firestoreApp struct
	firestoreApp := new(FirestoreApp)
	firestoreApp.Client = firestoreClient
	firestoreApp.Ctx = firebase.FirebaseInstance.Ctx

	return firestoreApp
}
