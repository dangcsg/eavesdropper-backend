package resources

import (
	"time"

	"cloud.google.com/go/firestore"
)

type TranscriptWhitelistedUser struct {
	ID         string // of this object
	Transcript *firestore.DocumentRef
	UserRef    *firestore.DocumentRef
	Handle     string
	Email      string
	CreatedAt  time.Time
}
