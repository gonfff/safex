package adapters

import (
	"fmt"

	"github.com/gonfff/safex/app/internal/opaqueauth"
)

// OpaqueAuthServiceAdapter adapter for OPAQUE authentication
type OpaqueAuthServiceAdapter struct {
	opaqueManager *opaqueauth.Manager
}

// NewOpaqueAuthServiceAdapter creates a new adapter for OPAQUE service
func NewOpaqueAuthServiceAdapter(opaqueManager *opaqueauth.Manager) *OpaqueAuthServiceAdapter {
	return &OpaqueAuthServiceAdapter{
		opaqueManager: opaqueManager,
	}
}

// RegistrationResponse creates a response for OPAQUE registration
func (s *OpaqueAuthServiceAdapter) RegistrationResponse(secretID string, payload []byte) ([]byte, error) {
	response, err := s.opaqueManager.RegistrationResponse(secretID, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration response: %w", err)
	}
	return response, nil
}

// LoginStart starts the OPAQUE login process
func (s *OpaqueAuthServiceAdapter) LoginStart(secretID string, opaqueRecord []byte, payload []byte) (sessionID string, response []byte, err error) {
	sessionID, response, err = s.opaqueManager.LoginStart(secretID, opaqueRecord, payload)
	if err != nil {
		// Convert opaqueauth errors to domain errors if needed
		return "", nil, fmt.Errorf("failed to start login: %w", err)
	}
	return sessionID, response, nil
}

// LoginFinish completes the OPAQUE login process
func (s *OpaqueAuthServiceAdapter) LoginFinish(sessionID string, finalization []byte) (secretID string, err error) {
	secretID, err = s.opaqueManager.LoginFinish(sessionID, finalization)
	if err != nil {
		// Can add conversion of specific opaqueauth errors to domain errors
		return "", fmt.Errorf("failed to finish login: %w", err)
	}
	return secretID, nil
}
