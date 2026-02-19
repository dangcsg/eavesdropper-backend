package whitelist

import (
	"context"
	"eavesdropper/dtos/requests"
	"eavesdropper/dtos/resources"
	"eavesdropper/dtos/responses"
	"eavesdropper/services/data/firestore/operations"
)

func AddUsersToTranscriptWhitelist(ctx context.Context, userID, transcriptID string, req *requests.AddToWhitelistRequest) error {
	whitelistedUsers := make([]resources.TranscriptWhitelistedUser, len(req.Users))

	for i, user := range req.Users {
		whitelistedUsers[i] = resources.TranscriptWhitelistedUser{
			Handle: user.Handle,
			Email:  user.Email,
		}
	}

	return operations.AddTranscriptWhitelistedUsers(ctx, userID, transcriptID, whitelistedUsers)
}

func RemoveUsersFromTranscriptWhitelist(ctx context.Context, userID, transcriptID string, req *requests.RemoveFromWhitelistRequest) error {
	return operations.RemoveTranscriptWhitelistedUsers(ctx, userID, transcriptID, req.Handles, req.Emails)
}

func GetTranscriptWhitelist(ctx context.Context, userID, transcriptID string) (*responses.WhitelistResponse, error) {
	whitelistedUsers, err := operations.GetTranscriptWhitelistedUsers(ctx, userID, transcriptID)
	if err != nil {
		return nil, err
	}

	responseUsers := make([]responses.WhitelistedUserResponse, len(whitelistedUsers))
	for i, whitelistedUser := range whitelistedUsers {
		// Get the actual user data from the user reference
		userDoc, err := whitelistedUser.UserRef.Get(ctx)
		if err != nil {
			return nil, err
		}

		user := new(resources.User)
		err = userDoc.DataTo(user)
		if err != nil {
			return nil, err
		}

		responseUsers[i] = responses.WhitelistedUserResponse{
			ID:        whitelistedUser.ID,
			Handle:    user.Handle,
			Email:     user.Email,
			CreatedAt: whitelistedUser.CreatedAt,
		}
	}

	return &responses.WhitelistResponse{
		Users: responseUsers,
	}, nil
}

func IsUserWhitelistedForTranscript(ctx context.Context, transcriptOwnerID, transcriptID, userID string) (bool, error) {
	return operations.IsUserWhitelistedForTranscript(ctx, transcriptOwnerID, transcriptID, userID)
}