package collections

import (
	cnfgs "eavesdropper/configurations"

	"cloud.google.com/go/firestore"
)

const transcriptWhitelistedUsersCollectionID = "transcriptWhitelistedUsers"
const transcriptWhitelistedUsersTestingCollectionID = "transcriptWhitelistedUsersTest"

var AllTranscriptWhitelistedUsers = getAllTranscriptWhitelistedUsersCollections()

func TranscriptWhitelistedUsers(userID, transcriptID string) *firestore.CollectionRef {
	collectionID := transcriptWhitelistedUsersTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = transcriptWhitelistedUsersCollectionID
	}
	return Transcripts(userID).Doc(transcriptID).Collection(collectionID)
}

func getAllTranscriptWhitelistedUsersCollections() *firestore.CollectionGroupRef {
	collectionID := transcriptWhitelistedUsersTestingCollectionID
	if cnfgs.SelectedBackendMode == cnfgs.Production {
		collectionID = transcriptWhitelistedUsersCollectionID
	}
	return db.CollectionGroup(collectionID)
}
