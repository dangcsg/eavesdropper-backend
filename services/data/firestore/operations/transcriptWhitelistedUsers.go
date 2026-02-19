package operations

import (
	"context"
	"eavesdropper/dtos/resources"
	"eavesdropper/services/data/firestore/collections"
	"errors"
	"fmt"
	"time"

	client "eavesdropper/services/data/firestore/client"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

var dbClient = client.FirestoreInstance.Client

func AddTranscriptWhitelistedUsers(
	ctx context.Context,
	userID string,
	transcriptID string,
	whitelistedUsers []resources.TranscriptWhitelistedUser,
) error {
	batch := dbClient.Batch()

	for _, user := range whitelistedUsers {
		// Find the actual user by handle or email
		var foundUserRef *firestore.DocumentRef
		var err error

		if user.Handle != "" {
			foundUserRef, err = findUserByHandle(ctx, user.Handle)
			if err != nil {
				return fmt.Errorf("failed to find user by handle %s: %w", user.Handle, err)
			}
		} else if user.Email != "" {
			foundUserRef, err = findUserByEmail(ctx, user.Email)
			if err != nil {
				return fmt.Errorf("failed to find user by email %s: %w", user.Email, err)
			}
		} else {
			return fmt.Errorf("either handle or email must be provided")
		}

		if foundUserRef == nil {
			return fmt.Errorf("user not found")
		}

		user.ID = uuid.NewString()
		user.CreatedAt = time.Now()
		user.Transcript = collections.Transcripts(userID).Doc(transcriptID)
		user.UserRef = foundUserRef

		docRef := collections.TranscriptWhitelistedUsers(userID, transcriptID).Doc(user.ID)
		batch.Create(docRef, user)
	}

	_, err := batch.Commit(ctx)
	return err
}

func RemoveTranscriptWhitelistedUsers(
	ctx context.Context,
	userID string,
	transcriptID string,
	handles []string,
	emails []string,
) error {
	batch := dbClient.Batch()

	// Remove by handles - find user references first
	for _, handle := range handles {
		userRef, err := findUserByHandle(ctx, handle)
		if err != nil {
			return fmt.Errorf("failed to find user by handle %s: %w", handle, err)
		}
		if userRef == nil {
			continue // Skip if user not found
		}

		iter := collections.TranscriptWhitelistedUsers(userID, transcriptID).
			Where("UserRef", "==", userRef).
			Documents(ctx)

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to find whitelisted user by user reference: %w", err)
			}
			batch.Delete(doc.Ref)
		}
	}

	// Remove by emails - find user references first
	for _, email := range emails {
		userRef, err := findUserByEmail(ctx, email)
		if err != nil {
			return fmt.Errorf("failed to find user by email %s: %w", email, err)
		}
		if userRef == nil {
			continue // Skip if user not found
		}

		iter := collections.TranscriptWhitelistedUsers(userID, transcriptID).
			Where("UserRef", "==", userRef).
			Documents(ctx)

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to find whitelisted user by user reference: %w", err)
			}
			batch.Delete(doc.Ref)
		}
	}

	_, err := batch.Commit(ctx)
	return err
}

func GetTranscriptWhitelistedUsers(ctx context.Context, userID, transcriptID string) ([]resources.TranscriptWhitelistedUser, error) {
	iter := collections.TranscriptWhitelistedUsers(userID, transcriptID).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	whitelistedUsers := []resources.TranscriptWhitelistedUser{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		user := new(resources.TranscriptWhitelistedUser)
		err = doc.DataTo(user)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to transfer data from snap to model: %s", err.Error()))
		}
		whitelistedUsers = append(whitelistedUsers, *user)
	}

	return whitelistedUsers, nil
}

func findUserByHandle(ctx context.Context, handle string) (*firestore.DocumentRef, error) {
	iter := collections.Users.Where("Handle", "==", handle).Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by handle: %w", err)
	}

	return doc.Ref, nil
}

func findUserByEmail(ctx context.Context, email string) (*firestore.DocumentRef, error) {
	iter := collections.Users.Where("Email", "==", email).Limit(1).Documents(ctx)

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // User not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	return doc.Ref, nil
}

func IsUserWhitelistedForTranscript(ctx context.Context, transcriptOwnerID, transcriptID, userID string) (bool, error) {
	userRef := collections.Users.Doc(userID)

	iter := collections.TranscriptWhitelistedUsers(transcriptOwnerID, transcriptID).
		Where("UserRef", "==", userRef).
		Limit(1).
		Documents(ctx)

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil // No matching whitelist entry found
	}
	if err != nil {
		return false, fmt.Errorf("failed to query whitelist: %w", err)
	}

	return true, nil // Found a matching whitelist entry
}
