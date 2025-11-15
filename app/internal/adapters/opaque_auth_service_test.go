package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOpaqueAuthServiceAdapter(t *testing.T) {
	// Test with nil manager
	adapter := NewOpaqueAuthServiceAdapter(nil)
	assert.NotNil(t, adapter)
	assert.Nil(t, adapter.opaqueManager)
}

func TestOpaqueAuthServiceAdapter_Methods_WithNilManager(t *testing.T) {
	adapter := NewOpaqueAuthServiceAdapter(nil)

	// Test RegistrationResponse with nil manager
	_, err := adapter.RegistrationResponse("test-id", []byte("payload"))
	assert.Error(t, err)

	// Test LoginStart with nil manager
	_, _, err = adapter.LoginStart("test-id", []byte("opaque-record"), []byte("payload"))
	assert.Error(t, err)

	// Test LoginFinish with nil manager
	_, err = adapter.LoginFinish("session-id", []byte("finalization"))
	assert.Error(t, err)
}
