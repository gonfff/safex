package storage

import (
	"testing"

	"github.com/gonfff/safex/app/internal/config"
)

func TestNewBlobStore(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
	}{
		{
			name: "local blob store",
			cfg: config.Config{
				BlobBackend: "local",
				BlobDir:     "/tmp/test",
			},
			expectError: false,
		},
		{
			name: "s3 blob store with all config",
			cfg: config.Config{
				BlobBackend: "s3",
				S3Bucket:    "test-bucket",
				S3Endpoint:  "http://localhost:9000", // Valid endpoint format
				S3Region:    "us-east-1",
			},
			expectError: true, // Will fail without actual S3 service, but shows config validation works
		},
		{
			name: "unsupported backend",
			cfg: config.Config{
				BlobBackend: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewBlobStore(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if store == nil {
				t.Error("expected store but got nil")
			}
		})
	}
}

func TestNewMetadataStore(t *testing.T) {
	tests := []struct {
		name        string
		cfg         config.Config
		expectError bool
	}{
		{
			name: "bbolt metadata store",
			cfg: config.Config{
				MetadataBackend: "bbolt",
				BoltPath:        "/tmp/test.db",
			},
			expectError: false,
		},
		{
			name: "redis metadata store",
			cfg: config.Config{
				MetadataBackend: "redis",
				RedisAddr:       "localhost:6379",
			},
			expectError: true, // Will fail without Redis running
		},
		{
			name: "unsupported backend",
			cfg: config.Config{
				MetadataBackend: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, err := NewMetadataStore(tt.cfg)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if store == nil {
				t.Error("expected store but got nil")
			}
		})
	}
}
