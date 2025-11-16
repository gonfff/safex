package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

func TestCleanupExpiredSecretsUseCase_Execute(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()

	tests := []struct {
		name        string
		setup       func(*mockSecretRepository, *mockBlobRepository, time.Time)
		expectError bool
		assert      func(*testing.T, *mockSecretRepository, *mockBlobRepository)
	}{
		{
			name: "no expired secrets",
			setup: func(sr *mockSecretRepository, br *mockBlobRepository, now time.Time) {
				sr.secrets["valid-id"] = &domain.Secret{
					ID:        "valid-id",
					ExpiresAt: now.Add(time.Hour),
				}
			},
			expectError: false,
			assert: func(t *testing.T, sr *mockSecretRepository, br *mockBlobRepository) {
				if len(sr.secrets) != 1 {
					t.Fatalf("expected secrets to remain untouched")
				}
			},
		},
		{
			name: "removes expired secrets",
			setup: func(sr *mockSecretRepository, br *mockBlobRepository, now time.Time) {
				sr.secrets["expired-id"] = &domain.Secret{
					ID:        "expired-id",
					ExpiresAt: now.Add(-time.Hour),
				}
				br.blobs["expired-id"] = []byte("payload")
			},
			expectError: false,
			assert: func(t *testing.T, sr *mockSecretRepository, br *mockBlobRepository) {
				if _, ok := sr.secrets["expired-id"]; ok {
					t.Fatalf("expected expired secret metadata to be removed")
				}
				if _, ok := br.blobs["expired-id"]; ok {
					t.Fatalf("expected expired secret blob to be removed")
				}
			},
		},
		{
			name: "list expired error",
			setup: func(sr *mockSecretRepository, br *mockBlobRepository, now time.Time) {
				sr.setListExpiredError(errors.New("list failed"))
			},
			expectError: true,
		},
		{
			name: "delete metadata error",
			setup: func(sr *mockSecretRepository, br *mockBlobRepository, now time.Time) {
				sr.secrets["expired-id"] = &domain.Secret{
					ID:        "expired-id",
					ExpiresAt: now.Add(-time.Hour),
				}
				sr.setDeleteError(errors.New("delete failed"))
			},
			expectError: true,
			assert: func(t *testing.T, sr *mockSecretRepository, br *mockBlobRepository) {
				if _, ok := sr.secrets["expired-id"]; !ok {
					t.Fatalf("expected secret metadata to remain due to delete error")
				}
			},
		},
		{
			name: "blob deletion error",
			setup: func(sr *mockSecretRepository, br *mockBlobRepository, now time.Time) {
				sr.secrets["expired-id"] = &domain.Secret{
					ID:        "expired-id",
					ExpiresAt: now.Add(-time.Hour),
				}
				br.blobs["expired-id"] = []byte("payload")
				br.setError(errors.New("blob delete failed"))
			},
			expectError: true,
			assert: func(t *testing.T, sr *mockSecretRepository, br *mockBlobRepository) {
				if _, ok := sr.secrets["expired-id"]; ok {
					t.Fatalf("expected metadata removed even if blob deletion failed")
				}
				if _, ok := br.blobs["expired-id"]; !ok {
					t.Fatalf("expected blob to remain due to deletion error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := newMockSecretRepository()
			br := newMockBlobRepository()
			now := time.Now()

			if tt.setup != nil {
				tt.setup(sr, br, now)
			}

			uc := NewCleanupExpiredSecretsUseCase(sr, br, logger)
			err := uc.Execute(ctx, now)

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.assert != nil {
				tt.assert(t, sr, br)
			}
		})
	}
}
