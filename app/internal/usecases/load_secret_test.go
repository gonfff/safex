package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

func TestLoadSecretUseCase_Execute(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()

	tests := []struct {
		name          string
		secretID      string
		setupMocks    func(*mockSecretRepository, *mockBlobRepository)
		expectError   bool
		expectedError error
	}{
		{
			name:     "successful load",
			secretID: "test-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				secret := &domain.Secret{
					ID:        "test-id",
					ExpiresAt: time.Now().Add(time.Hour),
				}
				sr.secrets["test-id"] = secret
				br.blobs["test-id"] = []byte("test payload")
			},
			expectError: false,
		},
		{
			name:     "secret not found",
			secretID: "nonexistent",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				// No setup needed - secret won't exist
			},
			expectError:   true,
			expectedError: domain.ErrSecretNotFound,
		},
		{
			name:     "expired secret",
			secretID: "expired-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				secret := &domain.Secret{
					ID:        "expired-id",
					ExpiresAt: time.Now().Add(-time.Hour),
				}
				sr.secrets["expired-id"] = secret
			},
			expectError:   true,
			expectedError: domain.ErrSecretExpired,
		},
		{
			name:     "blob retrieval error",
			secretID: "test-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				secret := &domain.Secret{
					ID:        "test-id",
					ExpiresAt: time.Now().Add(time.Hour),
				}
				sr.secrets["test-id"] = secret
				br.setError(errors.New("blob retrieval error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretRepo := newMockSecretRepository()
			blobRepo := newMockBlobRepository()
			tt.setupMocks(secretRepo, blobRepo)

			uc := NewLoadSecretUseCase(secretRepo, blobRepo, logger)
			result, err := uc.Execute(ctx, tt.secretID)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.expectedError != nil && err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result but got nil")
				return
			}

			if result.Payload == nil {
				t.Error("expected payload to be loaded")
			}
		})
	}
}

func TestLoadSecretUseCase_GetMetadata(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()

	tests := []struct {
		name          string
		secretID      string
		setupMocks    func(*mockSecretRepository, *mockBlobRepository)
		expectError   bool
		expectedError error
	}{
		{
			name:     "successful metadata load",
			secretID: "test-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				secret := &domain.Secret{
					ID:        "test-id",
					ExpiresAt: time.Now().Add(time.Hour),
				}
				sr.secrets["test-id"] = secret
			},
			expectError: false,
		},
		{
			name:     "secret not found",
			secretID: "nonexistent",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				// No setup needed
			},
			expectError:   true,
			expectedError: domain.ErrSecretNotFound,
		},
		{
			name:     "expired secret",
			secretID: "expired-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				secret := &domain.Secret{
					ID:        "expired-id",
					ExpiresAt: time.Now().Add(-time.Hour),
				}
				sr.secrets["expired-id"] = secret
			},
			expectError:   true,
			expectedError: domain.ErrSecretExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretRepo := newMockSecretRepository()
			blobRepo := newMockBlobRepository()
			tt.setupMocks(secretRepo, blobRepo)

			uc := NewLoadSecretUseCase(secretRepo, blobRepo, logger)
			result, err := uc.GetMetadata(ctx, tt.secretID)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.expectedError != nil && err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("expected result but got nil")
				return
			}

			// Metadata should not include payload
			if result.Payload != nil {
				t.Error("metadata should not include payload")
			}
		})
	}
}