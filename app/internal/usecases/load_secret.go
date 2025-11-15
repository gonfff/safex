package usecases

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

// LoadSecretUseCase use case for loading a secret
type LoadSecretUseCase struct {
	secretRepo domain.SecretRepository
	blobRepo   domain.BlobRepository
	logger     zerolog.Logger
}

// NewLoadSecretUseCase creates a new LoadSecretUseCase
func NewLoadSecretUseCase(
	secretRepo domain.SecretRepository,
	blobRepo domain.BlobRepository,
	logger zerolog.Logger,
) *LoadSecretUseCase {
	return &LoadSecretUseCase{
		secretRepo: secretRepo,
		blobRepo:   blobRepo,
		logger:     logger,
	}
}

// Execute performs secret loading with expiration check
func (uc *LoadSecretUseCase) Execute(ctx context.Context, id string) (*domain.Secret, error) {
	secret, err := uc.secretRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if secret.IsExpired() {
		// Delete expired secret
		_ = uc.deleteSecret(ctx, id)
		return nil, domain.ErrSecretExpired
	}

	// Load payload
	payload, err := uc.blobRepo.Retrieve(ctx, id)
	if err != nil {
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to retrieve blob")
		return nil, err
	}

	secret.Payload = payload
	return secret, nil
}

// GetMetadata returns only secret metadata without payload
func (uc *LoadSecretUseCase) GetMetadata(ctx context.Context, id string) (*domain.Secret, error) {
	secret, err := uc.secretRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if secret.IsExpired() {
		// Delete expired secret
		_ = uc.deleteSecret(ctx, id)
		return nil, domain.ErrSecretExpired
	}

	return secret, nil
}

func (uc *LoadSecretUseCase) deleteSecret(ctx context.Context, id string) error {
	if err := uc.secretRepo.Delete(ctx, id); err != nil {
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to delete expired secret metadata")
	}
	if err := uc.blobRepo.Remove(ctx, id); err != nil {
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to delete expired secret blob")
	}
	return nil
}
