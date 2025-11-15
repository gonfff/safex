package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/rs/zerolog"
)

func TestDeleteSecretUseCase_Execute(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()

	tests := []struct {
		name        string
		secretID    string
		setupMocks  func(*mockSecretRepository, *mockBlobRepository)
		expectError bool
	}{
		{
			name:     "successful deletion",
			secretID: "test-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				// No setup needed for successful case
			},
			expectError: false,
		},
		{
			name:     "secret deletion error",
			secretID: "test-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				sr.setError(errors.New("secret deletion error"))
			},
			expectError: true,
		},
		{
			name:     "blob deletion error",
			secretID: "test-id",
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				br.setError(errors.New("blob deletion error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretRepo := newMockSecretRepository()
			blobRepo := newMockBlobRepository()
			tt.setupMocks(secretRepo, blobRepo)

			uc := NewDeleteSecretUseCase(secretRepo, blobRepo, logger)
			err := uc.Execute(ctx, tt.secretID)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}