package errs

import (
	"errors"
)

// TODO RENAME TO ErrID prefix and set as string
var NoAudioChunksInManifest = errors.New("NoAudioChunksInManifest")
var ErrUserHasNoStripeAccount = errors.New("ErrUserHasNoStripeAccount")
var ErrUserHasNoActiveSubscription = errors.New("ErrUserHasNoActiveSubscription")
var ErrUserSubscriptionIsExpired = errors.New("ErrUserSubscriptionIsExpired")
var ErrExceededSubscriptionTranscriptionLimits = errors.New("ErrExceededSubscriptionTranscriptionLimits")
var ErrUnverifiedEmailAccount = errors.New("ErrUnverifiedEmailAccount")
var ErrInvalidAuthToken = errors.New("ErrInvalidAuthToken")
var ErrAuthTokenDoesNotMatchAcessedUser = errors.New("ErrAuthTokenDoesNotMatchAcessedUser")
