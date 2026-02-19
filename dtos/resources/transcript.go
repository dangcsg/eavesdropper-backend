package resources

import (
	"time"

	"cloud.google.com/go/firestore"
)

type Transcript struct {
	ID                        string
	UserRef                   *firestore.DocumentRef
	RecordingSessionID        string // mathces the manifest in storage
	Tittle                    string
	Content                   string
	ConsumedInputAudioSeconds int // total paid + free
	ConsumedFreeAudioSeconds  int
	ConsumedInputTokens       int // All input tokens including audio and insctructions
	ConsumedOutputTokens      int
	CreatedAt                 time.Time
	IsPrivate                 bool
}
