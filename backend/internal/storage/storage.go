package storage

import (
	"context"
	"fmt"

	"github.com/gonfff/safex/backend/internal/config"
	"github.com/gonfff/safex/backend/internal/storage/blob"
	"github.com/gonfff/safex/backend/internal/storage/metadata"
)

// BlobStore persists ciphertext payloads.
type BlobStore interface {
	Put(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// MetadataStore manages metadata, TTLs and etc.
type MetadataStore interface {
	Create(ctx context.Context, meta metadata.MetadataRecord) error
	Get(ctx context.Context, id string) (metadata.MetadataRecord, error)
	Delete(ctx context.Context, id string) error
}

// NewBlobStore initializes the configured blob store.
func NewBlobStore(cfg config.Config) (BlobStore, error) {
	switch cfg.BlobBackend {
	case "local":
		return blob.NewLocal(cfg.BlobDir)
	case "s3":
		return blob.NewS3(blob.S3Config{
			Bucket:    cfg.S3Bucket,
			Endpoint:  cfg.S3Endpoint,
			AccessKey: cfg.S3AccessKey,
			SecretKey: cfg.S3SecretKey,
			Region:    cfg.S3Region,
		})
	}
	return nil, fmt.Errorf("unsupported blob backend %q", cfg.BlobBackend)

}

// NewMetadataStore initializes metadata store according to config.
func NewMetadataStore(cfg config.Config) (MetadataStore, error) {
	switch cfg.MetadataBackend {
	case "bbolt":
		return metadata.NewBolt(cfg.BoltPath)
	case "redis":
		return metadata.NewRedis(metadata.RedisConfig{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
	}
	return nil, fmt.Errorf("unsupported metadata backend %q", cfg.MetadataBackend)

}
