package collections

import (
	cnfgs "eavesdropper/configurations"
	client "eavesdropper/services/data/firestore/client"

	"cloud.google.com/go/firestore"
)

var db = client.FirestoreInstance.Client

const usersCollectionID = "users"
const usersTestingCollectionID = "usersTest"

var Users = getUsersCollection()

func getUsersCollection() *firestore.CollectionRef {
	collectionID := usersTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = usersCollectionID
	}
	return db.Collection(collectionID)
}
