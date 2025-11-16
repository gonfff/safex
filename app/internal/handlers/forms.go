package handlers

import (
	"time"
)

// CreateSecretForm структура для валидации формы создания секрета
type CreateSecretForm struct {
	SecretID     string `form:"secret_id" binding:"required,uuid" validate:"required,uuid"`
	OpaqueUpload string `form:"opaque_upload" binding:"required,base64" validate:"required,base64"`
	TTLMinutes   int    `form:"ttl_minutes" binding:"omitempty,min=1,max=10080" validate:"omitempty,min=1,max=10080"` // max week
	Message      string `form:"message" binding:"omitempty,max=10485760" validate:"omitempty,max=10485760"`           // 10MB в символах
	PayloadType  string `form:"payload_type" binding:"omitempty,oneof=file text" validate:"omitempty,oneof=file text"`
}

// OpaqueRegisterStartForm структура для валидации регистрации
type OpaqueRegisterStartForm struct {
	Request string `json:"request" binding:"required,base64" validate:"required,base64"`
}

// OpaqueLoginStartForm структура для валидации входа
type OpaqueLoginStartForm struct {
	SecretID string `json:"secretId" binding:"required,uuid" validate:"required,uuid"`
	Request  string `json:"request" binding:"required,base64" validate:"required,base64"`
}

// RevealSecretForm структура для валидации раскрытия секрета
type RevealSecretForm struct {
	SessionID    string `form:"session_id" binding:"required,uuid" validate:"required,uuid"`
	Finalization string `form:"finalization" binding:"required,base64" validate:"required,base64"`
	SecretID     string `form:"secret_id" binding:"omitempty,uuid" validate:"omitempty,uuid"`
}

// GetTTLDuration возвращает время жизни секрета
func (f CreateSecretForm) GetTTLDuration(defaultTTL time.Duration) time.Duration {
	if f.TTLMinutes > 0 {
		return time.Duration(f.TTLMinutes) * time.Minute
	}
	return defaultTTL
}
