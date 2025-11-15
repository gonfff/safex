package opaqueauth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gonfff/safex/app/internal/config"
)

func TestNewManager_InvalidPrivateKey(t *testing.T) {
	cfg := config.Config{
		OpaquePrivateKey: "invalid-key",
		OpaqueOPRFSeed:   "dmFsaWQtb3ByZi1zZWVk", // "valid-oprf-seed" in base64
		OpaqueServerID:   "test-server",
		OpaqueSessionTTL: 120 * time.Second,
	}

	_, err := NewManager(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SAFEX_OPAQUE_PRIVATE_KEY")
}

func TestNewManager_InvalidOPRFSeed(t *testing.T) {
	cfg := config.Config{
		OpaquePrivateKey: "dmFsaWQtcHJpdmF0ZS1rZXk=", // "valid-private-key" in base64
		OpaqueOPRFSeed:   "invalid-seed",
		OpaqueServerID:   "test-server",
		OpaqueSessionTTL: 120 * time.Second,
	}

	_, err := NewManager(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "SAFEX_OPAQUE_OPRF_SEED")
}

func TestDecodeKey_ValidBase64(t *testing.T) {
	validKey := "aGVsbG8gd29ybGQ=" // "hello world" in base64

	decoded, err := decodeKey("TEST_KEY", validKey)
	assert.NoError(t, err)
	assert.Equal(t, []byte("hello world"), decoded)
}

func TestDecodeKey_InvalidBase64(t *testing.T) {
	invalidKey := "invalid-base64!"

	_, err := decodeKey("TEST_KEY", invalidKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_KEY")
}

func TestDecodeKey_EmptyKey(t *testing.T) {
	emptyKey := ""

	_, err := decodeKey("TEST_KEY", emptyKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TEST_KEY")
}
