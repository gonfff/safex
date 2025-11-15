package opaqueauth

/*
#cgo CFLAGS: -I${SRCDIR}/../../lib
#cgo LDFLAGS: -L${SRCDIR}/../../lib -lsafex_rust -ldl -lpthread
#include "safex_opaque.h"
#include <stdlib.h>
*/
import "C"

import (
	"encoding/base64"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"github.com/gonfff/safex/app/internal/config"
)

var (
	// ErrSessionNotFound indicates that the login session was not found.
	ErrSessionNotFound = errors.New("opaque session not found")
	// ErrSessionExpired indicates that the cached login session has expired.
	ErrSessionExpired = errors.New("opaque session expired")
)

// Manager holds a pointer to the Rust OPAQUE manager.
type Manager struct {
	ptr *C.SafexOpaqueManager
}

// NewManager initializes the Rust OPAQUE manager using application settings.
func NewManager(cfg config.Config) (*Manager, error) {
	secretKey, err := decodeKey("SAFEX_OPAQUE_PRIVATE_KEY", cfg.OpaquePrivateKey)
	if err != nil {
		return nil, err
	}
	oprfSeed, err := decodeKey("SAFEX_OPAQUE_OPRF_SEED", cfg.OpaqueOPRFSeed)
	if err != nil {
		return nil, err
	}
	serverID := []byte(strings.TrimSpace(cfg.OpaqueServerID))
	ttlSeconds := cfg.OpaqueSessionTTL.Seconds()
	if ttlSeconds <= 0 {
		ttlSeconds = 120
	}
	var errPtr *C.char
	manager := C.safex_opaque_manager_new(
		cBytes(serverID),
		C.size_t(len(serverID)),
		cBytes(secretKey),
		C.size_t(len(secretKey)),
		cBytes(oprfSeed),
		C.size_t(len(oprfSeed)),
		C.uint64_t(ttlSeconds),
		&errPtr,
	)
	if cerr := unwrapError(errPtr); cerr != nil {
		return nil, cerr
	}
	if manager == nil {
		return nil, errors.New("failed to initialize OPAQUE manager")
	}
	mgr := &Manager{ptr: manager}
	runtime.SetFinalizer(mgr, (*Manager).Close)
	return mgr, nil
}

// Close releases the underlying Rust manager, allowing Go's GC to reclaim it.
func (m *Manager) Close() {
	if m == nil || m.ptr == nil {
		return
	}
	C.safex_opaque_manager_free(m.ptr)
	m.ptr = nil
}

// RegistrationResponse computes the registration response for the provided secret ID.
func (m *Manager) RegistrationResponse(secretID string, payload []byte) ([]byte, error) {
	if m == nil || m.ptr == nil {
		return nil, errors.New("opaque manager not initialized")
	}
	var errPtr *C.char
	resp := C.safex_opaque_registration_response(
		m.ptr,
		cBytes([]byte(secretID)),
		C.size_t(len(secretID)),
		cBytes(payload),
		C.size_t(len(payload)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(resp)
	if cerr := unwrapError(errPtr); cerr != nil {
		return nil, mapError(cerr)
	}
	return copyBuffer(resp), nil
}

// LoginStart validates the KE1 message and returns KE2 plus a session ID.
func (m *Manager) LoginStart(secretID string, recordBlob []byte, ke1Payload []byte) (string, []byte, error) {
	if m == nil || m.ptr == nil {
		return "", nil, errors.New("opaque manager not initialized")
	}
	var errPtr *C.char
	result := C.safex_opaque_login_start(
		m.ptr,
		cBytes([]byte(secretID)),
		C.size_t(len(secretID)),
		cBytes(recordBlob),
		C.size_t(len(recordBlob)),
		cBytes(ke1Payload),
		C.size_t(len(ke1Payload)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(result.session_id)
	defer C.safex_opaque_buffer_free(result.response)
	if cerr := unwrapError(errPtr); cerr != nil {
		return "", nil, mapError(cerr)
	}
	sessionID := string(copyBuffer(result.session_id))
	ke2 := copyBuffer(result.response)
	return sessionID, ke2, nil
}

// LoginFinish verifies the KE3 payload for the given session and returns the secret ID.
func (m *Manager) LoginFinish(sessionID string, ke3Payload []byte) (string, error) {
	if m == nil || m.ptr == nil {
		return "", errors.New("opaque manager not initialized")
	}
	var errPtr *C.char
	resp := C.safex_opaque_login_finish(
		m.ptr,
		cBytes([]byte(sessionID)),
		C.size_t(len(sessionID)),
		cBytes(ke3Payload),
		C.size_t(len(ke3Payload)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(resp)
	if cerr := unwrapError(errPtr); cerr != nil {
		return "", mapError(cerr)
	}
	return string(copyBuffer(resp)), nil
}

func decodeKey(name, value string) ([]byte, error) {
	if strings.TrimSpace(value) == "" {
		return nil, fmt.Errorf("%s must be set", name)
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return decoded, nil
}

func cBytes(data []byte) *C.uint8_t {
	if len(data) == 0 {
		return nil
	}
	return (*C.uint8_t)(unsafe.Pointer(&data[0]))
}

func copyBuffer(buf C.SafexOpaqueBuffer) []byte {
	if buf.ptr == nil || buf.len == 0 {
		return nil
	}
	return C.GoBytes(unsafe.Pointer(buf.ptr), C.int(buf.len))
}

func unwrapError(errPtr *C.char) error {
	if errPtr == nil {
		return nil
	}
	defer C.safex_opaque_string_free(errPtr)
	return errors.New(C.GoString(errPtr))
}

func mapError(err error) error {
	switch err.Error() {
	case ErrSessionNotFound.Error():
		return ErrSessionNotFound
	case ErrSessionExpired.Error():
		return ErrSessionExpired
	}
	return err
}

// Client exposes Rust-backed helpers for driving the OPAQUE client flow in tests.
type Client struct{}

// NewClient constructs a helper capable of performing client-side OPAQUE steps.
func NewClient() *Client {
	return &Client{}
}

// StartRegistration begins the registration flow and returns a handle plus serialized message.
func (Client) StartRegistration(pin string) (uint32, []byte, error) {
	var errPtr *C.char
	result := C.safex_opaque_client_start_registration(
		cBytes([]byte(pin)),
		C.size_t(len(pin)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(result.message)
	if cerr := unwrapError(errPtr); cerr != nil {
		return 0, nil, cerr
	}
	return uint32(result.handle), copyBuffer(result.message), nil
}

// FinishRegistration finalizes the flow, producing the upload blob and export key.
func (Client) FinishRegistration(handle uint32, pin string, serverResponse []byte) ([]byte, []byte, error) {
	var errPtr *C.char
	result := C.safex_opaque_client_finish_registration(
		C.uint(handle),
		cBytes([]byte(pin)),
		C.size_t(len(pin)),
		cBytes(serverResponse),
		C.size_t(len(serverResponse)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(result.upload)
	defer C.safex_opaque_buffer_free(result.export_key)
	if cerr := unwrapError(errPtr); cerr != nil {
		return nil, nil, cerr
	}
	return copyBuffer(result.upload), copyBuffer(result.export_key), nil
}

// StartLogin begins the login flow and returns a handle plus the serialized credential request.
func (Client) StartLogin(pin string) (uint32, []byte, error) {
	var errPtr *C.char
	result := C.safex_opaque_client_start_login(
		cBytes([]byte(pin)),
		C.size_t(len(pin)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(result.message)
	if cerr := unwrapError(errPtr); cerr != nil {
		return 0, nil, cerr
	}
	return uint32(result.handle), copyBuffer(result.message), nil
}

// FinishLogin finalizes the login flow, returning the KE3 payload along with export/session keys.
func (Client) FinishLogin(handle uint32, pin string, serverResponse []byte) ([]byte, []byte, []byte, error) {
	var errPtr *C.char
	result := C.safex_opaque_client_finish_login(
		C.uint(handle),
		cBytes([]byte(pin)),
		C.size_t(len(pin)),
		cBytes(serverResponse),
		C.size_t(len(serverResponse)),
		&errPtr,
	)
	defer C.safex_opaque_buffer_free(result.finalization)
	defer C.safex_opaque_buffer_free(result.export_key)
	defer C.safex_opaque_buffer_free(result.session_key)
	if cerr := unwrapError(errPtr); cerr != nil {
		return nil, nil, nil, cerr
	}
	return copyBuffer(result.finalization), copyBuffer(result.export_key), copyBuffer(result.session_key), nil
}
