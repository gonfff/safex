package domain

import (
	"testing"
	"time"
)

func TestSecret_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{
			name:      "expired secret",
			expiresAt: time.Now().Add(-time.Hour),
			want:      true,
		},
		{
			name:      "not expired secret",
			expiresAt: time.Now().Add(time.Hour),
			want:      false,
		},
		{
			name:      "just expired",
			expiresAt: time.Now().Add(-time.Second),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Secret{
				ExpiresAt: tt.expiresAt,
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("Secret.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSecret_IsText(t *testing.T) {
	tests := []struct {
		name        string
		payloadType PayloadType
		want        bool
	}{
		{
			name:        "text payload",
			payloadType: PayloadTypeText,
			want:        true,
		},
		{
			name:        "file payload",
			payloadType: PayloadTypeFile,
			want:        false,
		},
		{
			name:        "empty payload type",
			payloadType: "",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Secret{
				PayloadType: tt.payloadType,
			}
			if got := s.IsText(); got != tt.want {
				t.Errorf("Secret.IsText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPayloadType_Constants(t *testing.T) {
	if PayloadTypeFile != "file" {
		t.Errorf("PayloadTypeFile = %v, want 'file'", PayloadTypeFile)
	}
	if PayloadTypeText != "text" {
		t.Errorf("PayloadTypeText = %v, want 'text'", PayloadTypeText)
	}
}
