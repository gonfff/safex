package handlers

import (
	"context"
	"errors"

	"github.com/gonfff/safex/app/internal/domain"
	"github.com/gonfff/safex/app/internal/usecases"
)

// Mock implementations for testing

type MockCreateSecretUseCase struct {
	ExecuteFunc func(ctx context.Context, input usecases.CreateSecretInput) (*domain.Secret, error)
}

func (m *MockCreateSecretUseCase) Execute(ctx context.Context, input usecases.CreateSecretInput) (*domain.Secret, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, input)
	}
	return &domain.Secret{
		ID:          input.ID,
		FileName:    input.FileName,
		ContentType: input.ContentType,
		Size:        int64(len(input.Payload)),
		PayloadType: input.PayloadType,
	}, nil
}

type MockLoadSecretUseCase struct {
	ExecuteFunc     func(ctx context.Context, id string) (*domain.Secret, error)
	GetMetadataFunc func(ctx context.Context, id string) (*domain.Secret, error)
}

func (m *MockLoadSecretUseCase) Execute(ctx context.Context, id string) (*domain.Secret, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, id)
	}
	return &domain.Secret{
		ID:       id,
		FileName: "test.txt",
		Payload:  []byte("test payload"),
	}, nil
}

func (m *MockLoadSecretUseCase) GetMetadata(ctx context.Context, id string) (*domain.Secret, error) {
	if m.GetMetadataFunc != nil {
		return m.GetMetadataFunc(ctx, id)
	}
	return &domain.Secret{
		ID:           id,
		FileName:     "test.txt",
		OpaqueRecord: []byte("opaque-record"),
	}, nil
}

type MockDeleteSecretUseCase struct {
	ExecuteFunc func(ctx context.Context, id string) error
}

func (m *MockDeleteSecretUseCase) Execute(ctx context.Context, id string) error {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, id)
	}
	return nil
}

type MockOpaqueAuthUseCase struct {
	StartRegistrationFunc func(secretID string, request []byte) ([]byte, error)
	StartLoginFunc        func(secretID string, opaqueRecord []byte, request []byte) (string, []byte, error)
	FinishLoginFunc       func(sessionID string, finalization []byte) (string, error)
}

func (m *MockOpaqueAuthUseCase) StartRegistration(secretID string, request []byte) ([]byte, error) {
	if m.StartRegistrationFunc != nil {
		return m.StartRegistrationFunc(secretID, request)
	}
	return []byte("opaque-response"), nil
}

func (m *MockOpaqueAuthUseCase) StartLogin(secretID string, opaqueRecord []byte, request []byte) (string, []byte, error) {
	if m.StartLoginFunc != nil {
		return m.StartLoginFunc(secretID, opaqueRecord, request)
	}
	return "session-id", []byte("login-response"), nil
}

func (m *MockOpaqueAuthUseCase) FinishLogin(sessionID string, finalization []byte) (string, error) {
	if m.FinishLoginFunc != nil {
		return m.FinishLoginFunc(sessionID, finalization)
	}
	return "secret-id", nil
}

// Common errors for testing
var (
	mockSecretNotFoundError = errors.New("secret not found")
	mockSecretExpiredError  = errors.New("secret expired")
	mockOpaqueError         = errors.New("opaque operation failed")
)
