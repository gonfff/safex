package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/backend/internal/storage"
	"github.com/gonfff/safex/backend/internal/storage/metadata"
)

// CreateInput captures the data required to persist a new secret.
type CreateInput struct {
	FileName    string
	ContentType string
	Payload     []byte
	TTL         time.Duration
}

// Service orchestrates metadata + blob stores.
type Service struct {
	blob   storage.BlobStore
	meta   storage.MetadataStore
	logger zerolog.Logger
}

// NewService builds a Service.
func NewService(blob storage.BlobStore, meta storage.MetadataStore, logger zerolog.Logger) *Service {
	return &Service{blob: blob, meta: meta, logger: logger}
}

// Create creates a new secret and returns its metadata.
func (s *Service) Create(ctx context.Context, input CreateInput) (metadata.MetadataRecord, error) {
	id := uuid.New().String()

	record := metadata.MetadataRecord{
		ID:          id,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		Size:        int64(len(input.Payload)),
		ExpiresAt:   time.Now().Add(input.TTL),
	}

	if err := s.blob.Put(ctx, id, input.Payload); err != nil {
		return metadata.MetadataRecord{}, fmt.Errorf("save blob: %w", err)
	}

	if err := s.meta.Create(ctx, record); err != nil {
		s.blob.Delete(ctx, id) // rollback blob on metadata failure
		return metadata.MetadataRecord{}, fmt.Errorf("store metadata: %w", err)
	}
	return record, nil
}

// Load loads metadata and payload, with expiration check.
func (s *Service) Load(ctx context.Context, id string) (metadata.MetadataRecord, []byte, error) {
	rec, err := s.meta.Get(ctx, id)
	if err != nil {
		return metadata.MetadataRecord{}, nil, fmt.Errorf("load metadata: %w", err)
	}
	if rec.ExpiresAt.Before(time.Now()) {
		_ = s.Delete(ctx, id)
		return metadata.MetadataRecord{}, nil, metadata.ErrExpired
	}
	payload, err := s.blob.Get(ctx, id)
	if err != nil {
		return metadata.MetadataRecord{}, nil, fmt.Errorf("load blob: %w", err)
	}
	return rec, payload, nil
}

// Delete removes both blob and metadata.
func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.meta.Delete(ctx, id); err != nil {
		return err
	}
	return s.blob.Delete(ctx, id)
}
