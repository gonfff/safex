package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_MaxPayloadBytes(t *testing.T) {
	cfg := Config{MaxPayloadMB: 5}
	expected := 5 * 1024 * 1024
	if got := cfg.MaxPayloadBytes(); got != expected {
		t.Errorf("MaxPayloadBytes() = %v, want %v", got, expected)
	}
}

func TestGetStrEnv(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal string
		envVal     string
		want       string
	}{
		{
			name:       "environment variable exists",
			key:        "TEST_STR",
			defaultVal: "default",
			envVal:     "env_value",
			want:       "env_value",
		},
		{
			name:       "environment variable does not exist",
			key:        "NON_EXISTENT",
			defaultVal: "default",
			want:       "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv(tt.key, tt.envVal)
				defer os.Unsetenv(tt.key)
			}

			got := getStrEnv(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("getStrEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		defaultVal int
		envVal     string
		want       int
	}{
		{
			name:       "valid integer environment variable",
			key:        "TEST_INT",
			defaultVal: 10,
			envVal:     "20",
			want:       20,
		},
		{
			name:       "invalid integer environment variable",
			key:        "TEST_INT_INVALID",
			defaultVal: 10,
			envVal:     "not_a_number",
			want:       10,
		},
		{
			name:       "environment variable does not exist",
			key:        "NON_EXISTENT_INT",
			defaultVal: 10,
			want:       10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVal != "" {
				os.Setenv(tt.key, tt.envVal)
				defer os.Unsetenv(tt.key)
			}

			got := getIntEnv(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("getIntEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_MustValidate(t *testing.T) {
	tests := []struct {
		name          string
		cfg           Config
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError: false,
		},
		{
			name: "invalid blob backend",
			cfg: Config{
				BlobBackend:       "invalid",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "unsupported blob backend",
		},
		{
			name: "invalid metadata backend",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "invalid",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "unsupported metadata backend",
		},
		{
			name: "s3 backend without bucket",
			cfg: Config{
				BlobBackend:       "s3",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_S3_BUCKET must be set",
		},
		{
			name: "s3 backend with partial credentials",
			cfg: Config{
				BlobBackend:       "s3",
				S3Bucket:          "test-bucket",
				S3AccessKey:       "access-key",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_S3_ACCESS_KEY and SAFEX_S3_SECRET_KEY must be set together",
		},
		{
			name: "redis backend without address",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "redis",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_REDIS_ADDR must be set",
		},
		{
			name: "zero max payload",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      0,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_MAX_PAYLOAD_MB must be greater than zero",
		},
		{
			name: "zero default TTL",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        0,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_DEFAULT_TTL_MINUTES must be greater than zero",
		},
		{
			name: "zero rate limit",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 0,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_RATE_LIMIT_PER_MINUTE must be greater than zero",
		},
		{
			name: "missing opaque private key",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_OPAQUE_PRIVATE_KEY must be set",
		},
		{
			name: "missing opaque OPRF seed",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueSessionTTL:  120 * time.Second,
			},
			expectError:   true,
			errorContains: "SAFEX_OPAQUE_OPRF_SEED must be set",
		},
		{
			name: "zero opaque session TTL",
			cfg: Config{
				BlobBackend:       "local",
				MetadataBackend:   "bbolt",
				MaxPayloadMB:      10,
				DefaultTTL:        15 * time.Minute,
				RequestsPerMinute: 10,
				OpaquePrivateKey:  "test-key",
				OpaqueOPRFSeed:    "test-seed",
				OpaqueSessionTTL:  0,
			},
			expectError:   true,
			errorContains: "SAFEX_OPAQUE_SESSION_TTL_SECONDS must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.MustValidate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Test loading with defaults (no environment variables set)
	cfg := Load()

	if cfg.HTTPAddr != ":8000" {
		t.Errorf("Expected default HTTPAddr :8000, got %s", cfg.HTTPAddr)
	}
	if cfg.MaxPayloadMB != 10 {
		t.Errorf("Expected default MaxPayloadMB 10, got %d", cfg.MaxPayloadMB)
	}
	if cfg.RequestsPerMinute != 10 {
		t.Errorf("Expected default RequestsPerMinute 10, got %d", cfg.RequestsPerMinute)
	}
	if cfg.BlobBackend != "local" {
		t.Errorf("Expected default BlobBackend 'local', got %s", cfg.BlobBackend)
	}
	if cfg.DefaultTTL != 15*time.Minute {
		t.Errorf("Expected default DefaultTTL 15 minutes, got %v", cfg.DefaultTTL)
	}

	// Verify that opaque keys are generated if not set
	if cfg.OpaquePrivateKey == "" {
		t.Error("Expected OpaquePrivateKey to be generated")
	}
	if cfg.OpaqueOPRFSeed == "" {
		t.Error("Expected OpaqueOPRFSeed to be generated")
	}
}

// containsString checks if a string contains a substring (case-sensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
