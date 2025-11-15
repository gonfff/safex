package usecases

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

// DeleteSecretUseCase use case for deleting a secret
type DeleteSecretUseCase struct {
	secretRepo domain.SecretRepository
	blobRepo   domain.BlobRepository
	logger     zerolog.Logger
}

// NewDeleteSecretUseCase creates a new DeleteSecretUseCase
func NewDeleteSecretUseCase(
	secretRepo domain.SecretRepository,
	blobRepo domain.BlobRepository,
	logger zerolog.Logger,
) *DeleteSecretUseCase {
	return &DeleteSecretUseCase{
		secretRepo: secretRepo,
		blobRepo:   blobRepo,
		logger:     logger,
	}
}

// Execute performs secret deletion
func (uc *DeleteSecretUseCase) Execute(ctx context.Context, id string) error {
	// Delete metadata first
	if err := uc.secretRepo.Delete(ctx, id); err != nil {
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to delete secret metadata")
		return err
	}

	// Then delete blob
	if err := uc.blobRepo.Remove(ctx, id); err != nil {
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to delete secret blob")
		return err
	}

	return nil
}
