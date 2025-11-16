package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

// CleanupExpiredSecretsUseCase removes secrets whose TTL elapsed.
type CleanupExpiredSecretsUseCase struct {
	secretRepo domain.SecretRepository
	blobRepo   domain.BlobRepository
	logger     zerolog.Logger
}

// NewCleanupExpiredSecretsUseCase constructs a cleanup use case instance.
func NewCleanupExpiredSecretsUseCase(
	secretRepo domain.SecretRepository,
	blobRepo domain.BlobRepository,
	logger zerolog.Logger,
) *CleanupExpiredSecretsUseCase {
	return &CleanupExpiredSecretsUseCase{
		secretRepo: secretRepo,
		blobRepo:   blobRepo,
		logger:     logger,
	}
}

// Execute removes all secrets that expired before the provided moment.
func (uc *CleanupExpiredSecretsUseCase) Execute(ctx context.Context, cutoff time.Time) error {
	expiredSecrets, err := uc.secretRepo.ListExpired(ctx, cutoff)
	if err != nil {
		return err
	}

	var result error
	for _, secret := range expiredSecrets {
		if err := uc.secretRepo.Delete(ctx, secret.ID); err != nil {
			uc.logger.Error().Err(err).Str("secret_id", secret.ID).Msg("cleanup: failed to delete metadata")
			result = errors.Join(result, err)
			continue
		}
		if err := uc.blobRepo.Remove(ctx, secret.ID); err != nil {
			uc.logger.Error().Err(err).Str("secret_id", secret.ID).Msg("cleanup: failed to delete blob")
			result = errors.Join(result, err)
			continue
		}
		uc.logger.Info().Str("secret_id", secret.ID).Msg("cleanup: expired secret removed")
	}

	return result
}
