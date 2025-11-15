package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

// CreateSecretInput input data for creating a secret
type CreateSecretInput struct {
	ID           string
	FileName     string
	ContentType  string
	Payload      []byte
	TTL          time.Duration
	PayloadType  domain.PayloadType
	OpaqueRecord []byte
}

// CreateSecretUseCase use case for creating a secret
type CreateSecretUseCase struct {
	secretRepo domain.SecretRepository
	blobRepo   domain.BlobRepository
	logger     zerolog.Logger
}

// NewCreateSecretUseCase creates a new CreateSecretUseCase
func NewCreateSecretUseCase(
	secretRepo domain.SecretRepository,
	blobRepo domain.BlobRepository,
	logger zerolog.Logger,
) *CreateSecretUseCase {
	return &CreateSecretUseCase{
		secretRepo: secretRepo,
		blobRepo:   blobRepo,
		logger:     logger,
	}
}

// Execute performs secret creation
func (uc *CreateSecretUseCase) Execute(ctx context.Context, input CreateSecretInput) (*domain.Secret, error) {
	if len(input.OpaqueRecord) == 0 {
		return nil, domain.ErrOpaqueRecordMissing
	}

	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = uuid.New().String()
	}

	payloadType := input.PayloadType
	if payloadType == "" {
		payloadType = domain.PayloadTypeFile
	}

	if input.TTL <= 0 {
		return nil, domain.ErrInvalidTTL
	}

	if len(input.Payload) == 0 {
		return nil, domain.ErrInvalidPayload
	}

	secret := &domain.Secret{
		ID:           id,
		FileName:     input.FileName,
		ContentType:  input.ContentType,
		Size:         int64(len(input.Payload)),
		ExpiresAt:    time.Now().Add(input.TTL),
		PayloadType:  payloadType,
		OpaqueRecord: input.OpaqueRecord,
		Payload:      input.Payload,
	}

	// Store blob first
	if err := uc.blobRepo.Store(ctx, id, input.Payload); err != nil {
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to store blob")
		return nil, err
	}

	// Then save metadata
	if err := uc.secretRepo.Create(ctx, secret); err != nil {
		// Rollback blob on metadata save error
		_ = uc.blobRepo.Remove(ctx, id)
		uc.logger.Error().Err(err).Str("secret_id", id).Msg("failed to create secret")
		return nil, err
	}

	// Don't return payload in result for security
	result := *secret
	result.Payload = nil
	return &result, nil
}
