package adapters

import (
	"context"
	"fmt"

	"github.com/gonfff/safex/app/internal/storage"
)

// BlobRepositoryAdapter adapter for blob storage
type BlobRepositoryAdapter struct {
	blobStore storage.BlobStore
}

// NewBlobRepositoryAdapter creates a new blob repository adapter
func NewBlobRepositoryAdapter(blobStore storage.BlobStore) *BlobRepositoryAdapter {
	return &BlobRepositoryAdapter{
		blobStore: blobStore,
	}
}

// Store saves data to blob storage
func (r *BlobRepositoryAdapter) Store(ctx context.Context, id string, data []byte) error {
	if err := r.blobStore.Put(ctx, id, data); err != nil {
		return fmt.Errorf("failed to store blob: %w", err)
	}
	return nil
}

// Retrieve extracts data from blob storage
func (r *BlobRepositoryAdapter) Retrieve(ctx context.Context, id string) ([]byte, error) {
	data, err := r.blobStore.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blob: %w", err)
	}
	return data, nil
}

// Remove deletes data from blob storage
func (r *BlobRepositoryAdapter) Remove(ctx context.Context, id string) error {
	if err := r.blobStore.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to remove blob: %w", err)
	}
	return nil
}
