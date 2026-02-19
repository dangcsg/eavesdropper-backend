package resources

// Represents a chunked audio file stored in storage.
// Stored by UI upon uploading audio to the storage.
// TODO Add duration in seconds (measure in UI and write there)
type RecordingManifest struct {
	UID       string   `json:"uid"`
	SessionID string   `json:"sessionId"`
	Chunks    []string `json:"chunks"`
	Count     int      `json:"count"`
	MIME      string   `json:"mime"`
}
