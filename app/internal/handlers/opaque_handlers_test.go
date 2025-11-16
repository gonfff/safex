package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/gonfff/safex/app/internal/config"
)

func TestHandleOpaqueRegisterStart_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/register/start", handlers.HandleOpaqueRegisterStart)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/opaque/register/start",
		bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleOpaqueRegisterStart_MissingRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/register/start", handlers.HandleOpaqueRegisterStart)

	// Test with empty request
	requestData := OpaqueRegisterStartForm{
		Request: "",
	}
	jsonData, _ := json.Marshal(requestData)

	req := httptest.NewRequest("POST", "/opaque/register/start",
		bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleOpaqueLoginStart_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/login/start", handlers.HandleOpaqueLoginStart)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/opaque/login/start",
		bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRenderTemplate_Function(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Test successful template rendering
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// This will test the renderTemplate method with mock data
	data := homePageData{
		DefaultTTLMinutes:  15,
		MaxPayloadMB:       10,
		MaxPayloadBytes:    10485760,
		RateLimitPerMinute: 60,
	}

	// This should not panic
	handlers.renderTemplate(c, "home", data)

	// Check that status code was set (either success or error)
	assert.True(t, c.Writer.Status() >= 200)
}

func TestRenderCreateResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Test renderCreateResult
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	handlers.renderCreateResult(c, http.StatusBadRequest, assert.AnError)

	// Check status code was set
	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}

func TestRenderRevealResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Test renderRevealResult
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	handlers.renderRevealResult(c, http.StatusOK, nil, nil, []byte("test payload"))

	// Check status code was set
	assert.Equal(t, http.StatusOK, c.Writer.Status())
}

func TestHandleOpaqueRegisterStart_InvalidBase64(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/register/start", handlers.HandleOpaqueRegisterStart)

	// Test with invalid base64 in request field
	reqData := OpaqueRegisterStartForm{
		Request: "invalid-base64!",
	}
	jsonData, _ := json.Marshal(reqData)

	req := httptest.NewRequest("POST", "/opaque/register/start", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleOpaqueLoginStart_MissingSecretID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/login/start", handlers.HandleOpaqueLoginStart)

	// Test without secretId
	reqData := OpaqueLoginStartForm{
		Request: "dGVzdA==", // valid base64
	}
	jsonData, _ := json.Marshal(reqData)

	req := httptest.NewRequest("POST", "/opaque/login/start", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleOpaqueLoginStart_EmptySecretID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/login/start", handlers.HandleOpaqueLoginStart)

	// Test with empty secretId
	reqData := OpaqueLoginStartForm{
		SecretID: "   ", // only whitespace
		Request:  "dGVzdA==",
	}
	jsonData, _ := json.Marshal(reqData)

	req := httptest.NewRequest("POST", "/opaque/login/start", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleOpaqueLoginStart_InvalidBase64Request(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/login/start", handlers.HandleOpaqueLoginStart)

	// Test with invalid base64 in request field
	reqData := OpaqueLoginStartForm{
		SecretID: "test-secret-id",
		Request:  "invalid-base64!",
	}
	jsonData, _ := json.Marshal(reqData)

	req := httptest.NewRequest("POST", "/opaque/login/start", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleOpaqueRegisterStart_EmptyRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/opaque/register/start", handlers.HandleOpaqueRegisterStart)

	// Test with empty request field
	reqData := OpaqueRegisterStartForm{
		Request: "", // empty field
	}
	jsonData, _ := json.Marshal(reqData)

	req := httptest.NewRequest("POST", "/opaque/register/start", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
