package operations

import (
	"context"
	"eavesdropper/dtos/resources"
	"eavesdropper/services/data/firestore/collections"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

func countUserTranscripts(ctx context.Context, userID string) (int64, error) {
	subcolRef := collections.Transcripts(userID)

	aggQuery := subcolRef.NewAggregationQuery().WithCount("total")
	aggSnap, err := aggQuery.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get aggregation: %w", err)
	}

	// raw, ok := aggSnap["total"]
	// fmt.Printf("\nGot raw count: %v, ok: %t\n", raw, ok)
	// fmt.Printf("Raw type: %v\n", reflect.TypeOf(raw))

	value, ok := aggSnap["total"].(*firestorepb.Value)
	if !ok {
		return 0, fmt.Errorf("failed to parse count")
	}

	return value.GetIntegerValue(), nil
}

func SaveTranscript(
	ctx context.Context,
	userID string,
	recordingSessionID string,
	transcriptStr string,
	consumedInputAudioSeconds int,
	consumedFreeAudioSeconds int,
	consumedInputTokens int,
	consumedOutputTokens int,
) (*resources.Transcript, error) {

	transcriptCount, err := countUserTranscripts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count existing transcripts: %w", err)
	}
	defaultTitle := fmt.Sprintf("Transcript #%d", transcriptCount+1)

	t := &resources.Transcript{
		ID:                        uuid.NewString(),
		UserRef:                   collections.Users.Doc(userID),
		RecordingSessionID:        recordingSessionID,
		Tittle:                    defaultTitle,
		Content:                   transcriptStr,
		ConsumedInputAudioSeconds: consumedInputAudioSeconds,
		ConsumedFreeAudioSeconds:  consumedFreeAudioSeconds,
		ConsumedInputTokens:       consumedInputTokens,
		ConsumedOutputTokens:      consumedOutputTokens,
		CreatedAt:                 time.Now(),
	}

	_, err = collections.Transcripts(userID).Doc(t.ID).Create(ctx, t)

	return t, err
}

func GetUserTranscripts(ctx context.Context, userID string, pageI int, pageSize int) ([]resources.Transcript, error) {
	query := collections.Transcripts(userID).
		OrderBy("CreatedAt", firestore.Desc)

	if pageI > 0 {
		query = query.Offset(pageI * pageSize)
	}

	if pageSize > 0 {
		query = query.Limit(pageSize)
	}

	iter := query.Documents(ctx)

	transcripts := []resources.Transcript{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		t := new(resources.Transcript)
		err = doc.DataTo(t)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to transfer data from snap to model: %s", err.Error()))
		}
		transcripts = append(transcripts, *t)
	}

	return transcripts, nil
}

func GetUserTranscript(ctx context.Context, userID, transcriptID string) (*resources.Transcript, error) {
	doc, err := collections.Transcripts(userID).Doc(transcriptID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript: %w", err)
	}

	transcript := new(resources.Transcript)
	err = doc.DataTo(transcript)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer data from document to model: %w", err)
	}

	return transcript, nil
}

func GetTranscripts(ctx context.Context, userID string, start, end time.Time) ([]resources.Transcript, error) {

	iter := collections.Transcripts(userID).
		Where("CreatedAt", ">=", start).
		Where("CreatedAt", "<=", end).
		OrderBy("CreatedAt", firestore.Desc).Documents(ctx)

	transcripts := []resources.Transcript{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		t := new(resources.Transcript)
		err = doc.DataTo(t)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to transfer data from snap to model: %s", err.Error()))
		}
		transcripts = append(transcripts, *t)
	}

	return transcripts, nil
}

func UpdateTranscriptPrivacy(ctx context.Context, userID, transcriptID string, isPrivate bool) error {
	_, err := collections.Transcripts(userID).Doc(transcriptID).Update(ctx, []firestore.Update{
		{Path: "IsPrivate", Value: isPrivate},
	})
	if err != nil {
		return fmt.Errorf("failed to update transcript privacy: %w", err)
	}
	return nil
}

func UpdateTranscriptTitle(ctx context.Context, userID, transcriptID, title string) error {
	_, err := collections.Transcripts(userID).Doc(transcriptID).Update(ctx, []firestore.Update{
		{Path: "Tittle", Value: title},
	})
	if err != nil {
		return fmt.Errorf("failed to update transcript title: %w", err)
	}
	return nil
}

func GetTranscriptByID(ctx context.Context, transcriptID string) (*resources.Transcript, string, error) {
	iter := collections.AllTranscripts.Where("ID", "==", transcriptID).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, "", fmt.Errorf("transcript not found")
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to query transcript: %w", err)
	}

	transcript := new(resources.Transcript)
	err = doc.DataTo(transcript)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse transcript: %w", err)
	}

	userID := doc.Ref.Parent.Parent.ID

	return transcript, userID, nil
}
