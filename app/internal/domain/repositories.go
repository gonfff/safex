package domain

import (
	"context"
	"time"
)

// SecretRepository interface for working with secrets
type SecretRepository interface {
	Create(ctx context.Context, secret *Secret) error
	GetByID(ctx context.Context, id string) (*Secret, error)
	Delete(ctx context.Context, id string) error
	ListExpired(ctx context.Context, before time.Time) ([]*Secret, error)
}

// BlobRepository interface for working with blob storage
type BlobRepository interface {
	Store(ctx context.Context, id string, data []byte) error
	Retrieve(ctx context.Context, id string) ([]byte, error)
	Remove(ctx context.Context, id string) error
}

// OpaqueAuthService interface for OPAQUE authentication
type OpaqueAuthService interface {
	RegistrationResponse(secretID string, payload []byte) ([]byte, error)
	LoginStart(secretID string, opaqueRecord []byte, payload []byte) (sessionID string, response []byte, err error)
	LoginFinish(sessionID string, finalization []byte) (secretID string, err error)
}
