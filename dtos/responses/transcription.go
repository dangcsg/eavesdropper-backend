package responses

import "time"

type TranscriptionResponse struct {
	ID                        string    `json:"id"`
	RecordingSessionID        string    `json:"recordingSessionID"`
	Tittle                    string    `json:"tittle"`
	Content                   string    `json:"content"`
	ConsumedInputAudioSeconds int       `json:"consumedInputAudioSeconds"`
	ConsumedFreeAudioSeconds  int       `json:"consumedFreeAudioSeconds"`
	CreatedAt                 time.Time `json:"createdAt"`
}
