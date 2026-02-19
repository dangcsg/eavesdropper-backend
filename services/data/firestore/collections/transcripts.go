package collections

import (
	cnfgs "eavesdropper/configurations"

	"cloud.google.com/go/firestore"
)

const transcriptsCollectionID = "transcripts"
const transcriptsTestingCollectionID = "transcriptsTest"

var AllTranscripts = getAllTranscriptsCollections()

func Transcripts(userID string) *firestore.CollectionRef {
	collectionID := transcriptsTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = transcriptsCollectionID
	}
	return Users.Doc(userID).Collection(collectionID)
}

func getAllTranscriptsCollections() *firestore.CollectionGroupRef {
	collectionID := transcriptsTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = transcriptsCollectionID
	}
	return db.CollectionGroup(collectionID)
}
