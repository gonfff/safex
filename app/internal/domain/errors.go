package domain

import "errors"

// Domain errors
var (
	ErrSecretNotFound      = errors.New("secret not found")
	ErrSecretExpired       = errors.New("secret expired")
	ErrSecretAlreadyExists = errors.New("secret already exists")
	ErrInvalidSecretID     = errors.New("invalid secret ID")
	ErrInvalidPayload      = errors.New("invalid payload")
	ErrInvalidTTL          = errors.New("invalid TTL")
	ErrOpaqueRecordMissing = errors.New("opaque record is required")
)
