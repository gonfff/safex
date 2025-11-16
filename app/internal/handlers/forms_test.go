package handlers

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestCreateSecretForm_Validation(t *testing.T) {
	validate := validator.New()
	// Регистрируем кастомные валидаторы
	validate.RegisterValidation("base64", validateBase64)
	validate.RegisterValidation("uuid", validateUUID)

	tests := []struct {
		name      string
		form      CreateSecretForm
		wantError bool
	}{
		{
			name: "valid form",
			form: CreateSecretForm{
				SecretID:     "550e8400-e29b-41d4-a716-446655440000",
				OpaqueUpload: "dGVzdA==", // "test" в base64
				TTLMinutes:   15,
				Message:      "test message",
				PayloadType:  "text",
			},
			wantError: false,
		},
		{
			name: "invalid UUID",
			form: CreateSecretForm{
				SecretID:     "not-a-uuid",
				OpaqueUpload: "dGVzdA==",
			},
			wantError: true,
		},
		{
			name: "invalid base64",
			form: CreateSecretForm{
				SecretID:     "550e8400-e29b-41d4-a716-446655440000",
				OpaqueUpload: "not-base64!",
			},
			wantError: true,
		},
		{
			name: "TTL too small",
			form: CreateSecretForm{
				SecretID:     "550e8400-e29b-41d4-a716-446655440000",
				OpaqueUpload: "dGVzdA==",
				TTLMinutes:   -1, // негативное значение должно вызвать ошибку
			},
			wantError: true,
		},
		{
			name: "invalid payload type",
			form: CreateSecretForm{
				SecretID:     "550e8400-e29b-41d4-a716-446655440000",
				OpaqueUpload: "dGVzdA==",
				PayloadType:  "invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Struct(&tt.form)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationMessages(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("base64", validateBase64)
	validate.RegisterValidation("uuid", validateUUID)

	form := CreateSecretForm{
		SecretID:     "invalid",
		OpaqueUpload: "",
	}

	err := validate.Struct(&form)
	assert.Error(t, err)

	validationErrors := FormatValidationErrors(err)
	assert.Len(t, validationErrors, 2)

	// Проверяем, что получили понятные сообщения
	var messages []string
	for _, e := range validationErrors {
		messages = append(messages, e.Message)
	}

	assert.Contains(t, messages, "SecretID must be a valid UUID")
	assert.Contains(t, messages, "OpaqueUpload is required")
}
