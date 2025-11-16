package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

// Mock implementations
type mockSecretRepository struct {
	secrets        map[string]*domain.Secret
	err            error
	deleteErr      error
	listExpiredErr error
}

func newMockSecretRepository() *mockSecretRepository {
	return &mockSecretRepository{
		secrets: make(map[string]*domain.Secret),
	}
}

func (m *mockSecretRepository) Create(ctx context.Context, secret *domain.Secret) error {
	if m.err != nil {
		return m.err
	}
	m.secrets[secret.ID] = secret
	return nil
}

func (m *mockSecretRepository) GetByID(ctx context.Context, id string) (*domain.Secret, error) {
	if m.err != nil {
		return nil, m.err
	}
	secret, exists := m.secrets[id]
	if !exists {
		return nil, domain.ErrSecretNotFound
	}
	return secret, nil
}

func (m *mockSecretRepository) Delete(ctx context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if m.err != nil {
		return m.err
	}
	delete(m.secrets, id)
	return nil
}

func (m *mockSecretRepository) ListExpired(ctx context.Context, before time.Time) ([]*domain.Secret, error) {
	if m.listExpiredErr != nil {
		return nil, m.listExpiredErr
	}
	if m.err != nil {
		return nil, m.err
	}
	expired := make([]*domain.Secret, 0)
	for _, secret := range m.secrets {
		if !secret.ExpiresAt.After(before) {
			expired = append(expired, secret)
		}
	}
	return expired, nil
}

func (m *mockSecretRepository) setError(err error) {
	m.err = err
}

func (m *mockSecretRepository) setDeleteError(err error) {
	m.deleteErr = err
}

func (m *mockSecretRepository) setListExpiredError(err error) {
	m.listExpiredErr = err
}

type mockBlobRepository struct {
	blobs map[string][]byte
	err   error
}

func newMockBlobRepository() *mockBlobRepository {
	return &mockBlobRepository{
		blobs: make(map[string][]byte),
	}
}

func (m *mockBlobRepository) Store(ctx context.Context, id string, data []byte) error {
	if m.err != nil {
		return m.err
	}
	m.blobs[id] = data
	return nil
}

func (m *mockBlobRepository) Retrieve(ctx context.Context, id string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	data, exists := m.blobs[id]
	if !exists {
		return nil, errors.New("blob not found")
	}
	return data, nil
}

func (m *mockBlobRepository) Remove(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.blobs, id)
	return nil
}

func (m *mockBlobRepository) setError(err error) {
	m.err = err
}

func TestCreateSecretUseCase_Execute(t *testing.T) {
	logger := zerolog.Nop()
	ctx := context.Background()

	tests := []struct {
		name          string
		input         CreateSecretInput
		setupMocks    func(*mockSecretRepository, *mockBlobRepository)
		expectError   bool
		expectedError error
	}{
		{
			name: "successful creation",
			input: CreateSecretInput{
				ID:           "test-id",
				FileName:     "test.txt",
				ContentType:  "text/plain",
				Payload:      []byte("test payload"),
				TTL:          time.Hour,
				PayloadType:  domain.PayloadTypeText,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks:  func(sr *mockSecretRepository, br *mockBlobRepository) {},
			expectError: false,
		},
		{
			name: "missing opaque record",
			input: CreateSecretInput{
				ID:          "test-id",
				Payload:     []byte("test payload"),
				TTL:         time.Hour,
				PayloadType: domain.PayloadTypeText,
			},
			setupMocks:    func(sr *mockSecretRepository, br *mockBlobRepository) {},
			expectError:   true,
			expectedError: domain.ErrOpaqueRecordMissing,
		},
		{
			name: "invalid TTL",
			input: CreateSecretInput{
				ID:           "test-id",
				Payload:      []byte("test payload"),
				TTL:          0,
				PayloadType:  domain.PayloadTypeText,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks:    func(sr *mockSecretRepository, br *mockBlobRepository) {},
			expectError:   true,
			expectedError: domain.ErrInvalidTTL,
		},
		{
			name: "invalid payload",
			input: CreateSecretInput{
				ID:           "test-id",
				TTL:          time.Hour,
				PayloadType:  domain.PayloadTypeText,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks:    func(sr *mockSecretRepository, br *mockBlobRepository) {},
			expectError:   true,
			expectedError: domain.ErrInvalidPayload,
		},
		{
			name: "blob store error",
			input: CreateSecretInput{
				ID:           "test-id",
				Payload:      []byte("test payload"),
				TTL:          time.Hour,
				PayloadType:  domain.PayloadTypeText,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				br.setError(errors.New("blob store error"))
			},
			expectError: true,
		},
		{
			name: "secret create error with rollback",
			input: CreateSecretInput{
				ID:           "test-id",
				Payload:      []byte("test payload"),
				TTL:          time.Hour,
				PayloadType:  domain.PayloadTypeText,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks: func(sr *mockSecretRepository, br *mockBlobRepository) {
				// First call (store) succeeds, second call (create) fails
				sr.setError(errors.New("secret create error"))
			},
			expectError: true,
		},
		{
			name: "auto-generated ID",
			input: CreateSecretInput{
				ID:           "",
				Payload:      []byte("test payload"),
				TTL:          time.Hour,
				PayloadType:  domain.PayloadTypeText,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks:  func(sr *mockSecretRepository, br *mockBlobRepository) {},
			expectError: false,
		},
		{
			name: "default payload type",
			input: CreateSecretInput{
				ID:           "test-id",
				Payload:      []byte("test payload"),
				TTL:          time.Hour,
				OpaqueRecord: []byte("opaque-record"),
			},
			setupMocks:  func(sr *mockSecretRepository, br *mockBlobRepository) {},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secretRepo := newMockSecretRepository()
			blobRepo := newMockBlobRepository()
			tt.setupMocks(secretRepo, blobRepo)

			uc := NewCreateSecretUseCase(secretRepo, blobRepo, logger)
			result, err := uc.Execute(ctx, tt.input)

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

			// Check that payload is not returned in result
			if result.Payload != nil {
				t.Error("payload should not be returned in result")
			}

			// Verify secret was stored
			if len(secretRepo.secrets) != 1 {
				t.Errorf("expected 1 secret in repository, got %d", len(secretRepo.secrets))
			}

			// Verify blob was stored
			if len(blobRepo.blobs) != 1 {
				t.Errorf("expected 1 blob in repository, got %d", len(blobRepo.blobs))
			}
		})
	}
}
