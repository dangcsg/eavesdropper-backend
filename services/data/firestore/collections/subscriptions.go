package collections

import (
	cnfgs "eavesdropper/configurations"

	"cloud.google.com/go/firestore"
)

// // Subscriptions (Subcollection of Users) ////
const subscriptionsCollectionID = "subscriptions"
const subscriptionsTestingCollectionID = "subscriptionsTest"

var AllSubscriptions = getAllSubscriptionsCollections()

func Subscriptions(userID string) *firestore.CollectionRef {
	collectionID := subscriptionsTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = subscriptionsCollectionID
	}
	return Users.Doc(userID).Collection(collectionID)
}

func getAllSubscriptionsCollections() *firestore.CollectionGroupRef {
	collectionID := subscriptionsTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = subscriptionsCollectionID
	}
	return db.CollectionGroup(collectionID)
}
