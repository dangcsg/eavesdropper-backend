package transcripts

import (
	"context"
	"eavesdropper/dtos/resources"
	"eavesdropper/services/data/firestore/operations"
	db "eavesdropper/services/data/firestore/operations"

	"google.golang.org/genai"
)

func SaveTranscript(
	ctx context.Context,
	genaiResponse *genai.GenerateContentResponse,
	userID,
	recordingSessionID string,
	audioSeconds, consumedFreeAudioSeconds int,
) (*resources.Transcript, error) {

	transcript, err := operations.SaveTranscript(
		ctx,
		userID,
		recordingSessionID,
		genaiResponse.Text(),
		audioSeconds,
		consumedFreeAudioSeconds,
		int(genaiResponse.UsageMetadata.PromptTokenCount),
		int(genaiResponse.UsageMetadata.CandidatesTokenCount),
	)

	return transcript, err
}

func GetUserTranscripts(ctx context.Context, userID string, pageI int, pageSize int) ([]resources.Transcript, error) {
	return db.GetUserTranscripts(ctx, userID, pageI, pageSize)
}

func GetUserTranscript(ctx context.Context, userID, transcriptID string) (*resources.Transcript, error) {
	return db.GetUserTranscript(ctx, userID, transcriptID)
}

func UpdateTranscriptPrivacy(ctx context.Context, userID, transcriptID string, isPrivate bool) error {
	return db.UpdateTranscriptPrivacy(ctx, userID, transcriptID, isPrivate)
}

func UpdateTranscriptTitle(ctx context.Context, userID, transcriptID, title string) error {
	return db.UpdateTranscriptTitle(ctx, userID, transcriptID, title)
}

func GetTranscriptByID(ctx context.Context, transcriptID string) (*resources.Transcript, string, error) {
	return db.GetTranscriptByID(ctx, transcriptID)
}
