package auth

import (
	"context"
	"eavesdropper/errs"

	"firebase.google.com/go/v4/auth"
)

func ValidateToken(ctx context.Context, tokenID string) error {
	_, err := Auth.Client.VerifyIDToken(ctx, tokenID)
	return err
}

// TODO CAN TOKEN BE NULL WIHTOUT ERROR?
func GetUserID(ctx context.Context, tokenID string) (string, error) {

	token, err := Auth.Client.VerifyIDToken(ctx, tokenID)
	if err != nil {
		return "", err
	}

	return token.UID, nil
}

func GetUserRecord(ctx context.Context, tokenID string) (*auth.UserRecord, error) {
	token, err := Auth.Client.VerifyIDToken(ctx, tokenID)
	if err != nil {
		return nil, err
	}

	user, err := Auth.Client.GetUser(ctx, token.UID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func GetVerifiedUserID(ctx context.Context, tokenID string) (string, error) {

	token, err := Auth.Client.VerifyIDToken(ctx, tokenID)
	if err != nil {
		return "", err
	}

	user, err := Auth.Client.GetUser(ctx, token.UID)
	if err != nil {
		return "", err
	}
	if !user.EmailVerified {
		return "", errs.ErrUnverifiedEmailAccount
	}

	return token.UID, nil
}
