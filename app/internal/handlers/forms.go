package handlers

import (
	"errors"
	"time"

	"github.com/gonfff/safex/app/internal/config"
)

// CreateSecretForm struct for validating secret creation
type CreateSecretForm struct {
	SecretID     string `form:"secret_id" binding:"required,uuid" validate:"required,uuid"`
	OpaqueUpload string `form:"opaque_upload" binding:"required,base64" validate:"required,base64"`
	TTLMinutes   int    `form:"ttl_minutes" binding:"omitempty,min=1" validate:"omitempty,min=1"`
	Message      string `form:"message" binding:"omitempty" validate:"omitempty"`
	PayloadType  string `form:"payload_type" binding:"omitempty,oneof=file text" validate:"omitempty,oneof=file text"`
}

// ValidateWithConfig validates the form using configuration limits
func (f *CreateSecretForm) ValidateWithConfig(cfg config.Config) error {
	if f.TTLMinutes > cfg.MaxTTLMinutes {
		return errors.New("ttl_minutes exceeds maximum allowed value")
	}

	if len(f.Message) > cfg.MaxPayloadBytes() {
		return errors.New("message size exceeds maximum allowed payload size")
	}

	return nil
} // OpaqueRegisterStartForm struct for validating registration
type OpaqueRegisterStartForm struct {
	Request string `json:"request" binding:"required,base64" validate:"required,base64"`
}

// OpaqueLoginStartForm struct for validating login
type OpaqueLoginStartForm struct {
	SecretID string `json:"secretId" binding:"required,uuid" validate:"required,uuid"`
	Request  string `json:"request" binding:"required,base64" validate:"required,base64"`
}

// RevealSecretForm struct for validating secret revelation
type RevealSecretForm struct {
	SessionID    string `form:"session_id" binding:"required,uuid" validate:"required,uuid"`
	Finalization string `form:"finalization" binding:"required,base64" validate:"required,base64"`
	SecretID     string `form:"secret_id" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

// GetTTLDuration returns the secret's time-to-live duration
func (f CreateSecretForm) GetTTLDuration(defaultTTL time.Duration) time.Duration {
	if f.TTLMinutes > 0 {
		return time.Duration(f.TTLMinutes) * time.Minute
	}
	return defaultTTL
}
