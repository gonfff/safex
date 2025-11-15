package usecases

import (
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/domain"
)

// OpaqueAuthUseCase use case for OPAQUE authentication
type OpaqueAuthUseCase struct {
	opaqueService domain.OpaqueAuthService
	logger        zerolog.Logger
}

// NewOpaqueAuthUseCase creates a new OpaqueAuthUseCase
func NewOpaqueAuthUseCase(
	opaqueService domain.OpaqueAuthService,
	logger zerolog.Logger,
) *OpaqueAuthUseCase {
	return &OpaqueAuthUseCase{
		opaqueService: opaqueService,
		logger:        logger,
	}
}

// StartRegistration starts the OPAQUE registration process
func (uc *OpaqueAuthUseCase) StartRegistration(secretID string, payload []byte) ([]byte, error) {
	response, err := uc.opaqueService.RegistrationResponse(secretID, payload)
	if err != nil {
		uc.logger.Error().Err(err).Str("secret_id", secretID).Msg("failed to start opaque registration")
		return nil, err
	}
	return response, nil
}

// StartLogin starts the OPAQUE login process
func (uc *OpaqueAuthUseCase) StartLogin(secretID string, opaqueRecord []byte, payload []byte) (sessionID string, response []byte, err error) {
	sessionID, response, err = uc.opaqueService.LoginStart(secretID, opaqueRecord, payload)
	if err != nil {
		uc.logger.Error().Err(err).Str("secret_id", secretID).Msg("failed to start opaque login")
		return "", nil, err
	}
	return sessionID, response, nil
}

// FinishLogin completes the OPAQUE login process
func (uc *OpaqueAuthUseCase) FinishLogin(sessionID string, finalization []byte) (secretID string, err error) {
	secretID, err = uc.opaqueService.LoginFinish(sessionID, finalization)
	if err != nil {
		uc.logger.Error().Err(err).Str("session_id", sessionID).Msg("failed to finish opaque login")
		return "", err
	}
	return secretID, nil
}
