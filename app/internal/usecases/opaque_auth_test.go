package usecases

import (
	"errors"
	"testing"

	"github.com/rs/zerolog"
)

// Mock OpaqueAuthService
type mockOpaqueAuthService struct {
	err error
}

func newMockOpaqueAuthService() *mockOpaqueAuthService {
	return &mockOpaqueAuthService{}
}

func (m *mockOpaqueAuthService) RegistrationResponse(secretID string, payload []byte) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte("registration-response"), nil
}

func (m *mockOpaqueAuthService) LoginStart(secretID string, opaqueRecord []byte, payload []byte) (sessionID string, response []byte, err error) {
	if m.err != nil {
		return "", nil, m.err
	}
	return "session-id", []byte("login-response"), nil
}

func (m *mockOpaqueAuthService) LoginFinish(sessionID string, finalization []byte) (secretID string, err error) {
	if m.err != nil {
		return "", m.err
	}
	return "secret-id", nil
}

func (m *mockOpaqueAuthService) setError(err error) {
	m.err = err
}

func TestOpaqueAuthUseCase_StartRegistration(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		name        string
		secretID    string
		payload     []byte
		setupMock   func(*mockOpaqueAuthService)
		expectError bool
	}{
		{
			name:        "successful registration",
			secretID:    "test-id",
			payload:     []byte("test-payload"),
			setupMock:   func(m *mockOpaqueAuthService) {},
			expectError: false,
		},
		{
			name:     "registration error",
			secretID: "test-id",
			payload:  []byte("test-payload"),
			setupMock: func(m *mockOpaqueAuthService) {
				m.setError(errors.New("registration error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockOpaqueAuthService()
			tt.setupMock(mockService)

			uc := NewOpaqueAuthUseCase(mockService, logger)
			result, err := uc.StartRegistration(tt.secretID, tt.payload)

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

			if result == nil {
				t.Error("expected result but got nil")
			}
		})
	}
}

func TestOpaqueAuthUseCase_StartLogin(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		name         string
		secretID     string
		opaqueRecord []byte
		payload      []byte
		setupMock    func(*mockOpaqueAuthService)
		expectError  bool
	}{
		{
			name:         "successful login start",
			secretID:     "test-id",
			opaqueRecord: []byte("opaque-record"),
			payload:      []byte("test-payload"),
			setupMock:    func(m *mockOpaqueAuthService) {},
			expectError:  false,
		},
		{
			name:         "login start error",
			secretID:     "test-id",
			opaqueRecord: []byte("opaque-record"),
			payload:      []byte("test-payload"),
			setupMock: func(m *mockOpaqueAuthService) {
				m.setError(errors.New("login start error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockOpaqueAuthService()
			tt.setupMock(mockService)

			uc := NewOpaqueAuthUseCase(mockService, logger)
			sessionID, response, err := uc.StartLogin(tt.secretID, tt.opaqueRecord, tt.payload)

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

			if sessionID == "" {
				t.Error("expected sessionID but got empty string")
			}

			if response == nil {
				t.Error("expected response but got nil")
			}
		})
	}
}

func TestOpaqueAuthUseCase_FinishLogin(t *testing.T) {
	logger := zerolog.Nop()

	tests := []struct {
		name         string
		sessionID    string
		finalization []byte
		setupMock    func(*mockOpaqueAuthService)
		expectError  bool
	}{
		{
			name:         "successful login finish",
			sessionID:    "session-id",
			finalization: []byte("finalization"),
			setupMock:    func(m *mockOpaqueAuthService) {},
			expectError:  false,
		},
		{
			name:         "login finish error",
			sessionID:    "session-id",
			finalization: []byte("finalization"),
			setupMock: func(m *mockOpaqueAuthService) {
				m.setError(errors.New("login finish error"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := newMockOpaqueAuthService()
			tt.setupMock(mockService)

			uc := NewOpaqueAuthUseCase(mockService, logger)
			secretID, err := uc.FinishLogin(tt.sessionID, tt.finalization)

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

			if secretID == "" {
				t.Error("expected secretID but got empty string")
			}
		})
	}
}
