package handlers

import (
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/gonfff/safex/app/internal/config"
)

func TestCreateSecretForm_Validation(t *testing.T) {
	validate := validator.New()
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
				OpaqueUpload: "dGVzdA==", // "test" in base64
				TTLValue:     2,
				TTLUnit:      "hours",
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
				TTLValue:     -1,
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
		{
			name: "invalid ttl unit",
			form: CreateSecretForm{
				SecretID:     "550e8400-e29b-41d4-a716-446655440000",
				OpaqueUpload: "dGVzdA==",
				TTLValue:     10,
				TTLUnit:      "weeks",
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

	// Check that we received clear messages
	var messages []string
	for _, e := range validationErrors {
		messages = append(messages, e.Message)
	}

	assert.Contains(t, messages, "SecretID must be a valid UUID")
	assert.Contains(t, messages, "OpaqueUpload is required")
}

func TestCreateSecretForm_ValidateWithConfig_TTLConversion(t *testing.T) {
	cfg := config.Config{MaxTTLMinutes: 24 * 60}

	form := CreateSecretForm{
		TTLValue: 1,
		TTLUnit:  "days",
	}
	assert.NoError(t, form.ValidateWithConfig(cfg))

	tooLarge := CreateSecretForm{
		TTLValue: 2,
		TTLUnit:  "days",
	}
	assert.EqualError(
		t,
		tooLarge.ValidateWithConfig(cfg),
		"ttl exceeds maximum allowed value",
	)

	legacy := CreateSecretForm{
		TTLMinutes: 15,
	}
	assert.NoError(t, legacy.ValidateWithConfig(cfg))
}

func TestCreateSecretForm_GetTTLDuration(t *testing.T) {
	defaultTTL := 15 * time.Minute

	tests := []struct {
		name string
		form CreateSecretForm
		want time.Duration
	}{
		{
			name: "legacy ttl minutes",
			form: CreateSecretForm{TTLMinutes: 30},
			want: 30 * time.Minute,
		},
		{
			name: "hours unit",
			form: CreateSecretForm{TTLValue: 2, TTLUnit: "hours"},
			want: 2 * time.Hour,
		},
		{
			name: "days unit",
			form: CreateSecretForm{TTLValue: 1, TTLUnit: "days"},
			want: 24 * time.Hour,
		},
		{
			name: "fallback to default when ttl missing",
			form: CreateSecretForm{},
			want: defaultTTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.form.GetTTLDuration(defaultTTL))
		})
	}
}
