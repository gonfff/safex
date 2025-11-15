package adapters

import (
	"context"
	"fmt"

	"github.com/gonfff/safex/app/internal/domain"
	"github.com/gonfff/safex/app/internal/storage"
	"github.com/gonfff/safex/app/internal/storage/metadata"
)

// SecretRepositoryAdapter adapter for secret storage
type SecretRepositoryAdapter struct {
	metaStore storage.MetadataStore
}

// NewSecretRepositoryAdapter creates a new secret repository adapter
func NewSecretRepositoryAdapter(metaStore storage.MetadataStore) *SecretRepositoryAdapter {
	return &SecretRepositoryAdapter{
		metaStore: metaStore,
	}
}

// Create creates a secret in storage
func (r *SecretRepositoryAdapter) Create(ctx context.Context, secret *domain.Secret) error {
	record := metadata.MetadataRecord{
		ID:           secret.ID,
		FileName:     secret.FileName,
		ContentType:  secret.ContentType,
		Size:         secret.Size,
		ExpiresAt:    secret.ExpiresAt,
		PayloadType:  metadata.PayloadType(secret.PayloadType),
		OpaqueRecord: secret.OpaqueRecord,
	}

	if err := r.metaStore.Create(ctx, record); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	return nil
}

// GetByID retrieves a secret by ID
func (r *SecretRepositoryAdapter) GetByID(ctx context.Context, id string) (*domain.Secret, error) {
	record, err := r.metaStore.Get(ctx, id)
	if err != nil {
		if err == metadata.ErrNotFound {
			return nil, domain.ErrSecretNotFound
		}
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	secret := &domain.Secret{
		ID:           record.ID,
		FileName:     record.FileName,
		ContentType:  record.ContentType,
		Size:         record.Size,
		ExpiresAt:    record.ExpiresAt,
		PayloadType:  domain.PayloadType(record.PayloadType),
		OpaqueRecord: record.OpaqueRecord,
	}

	return secret, nil
}

// Delete removes a secret by ID
func (r *SecretRepositoryAdapter) Delete(ctx context.Context, id string) error {
	if err := r.metaStore.Delete(ctx, id); err != nil {
		if err == metadata.ErrNotFound {
			return domain.ErrSecretNotFound
		}
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}
