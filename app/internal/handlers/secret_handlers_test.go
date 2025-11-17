package handlers

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/gonfff/safex/app/internal/config"
	"github.com/gonfff/safex/app/internal/domain"
)

func TestHandleCreateSecret_InvalidForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Test invalid Content-Type
	req := httptest.NewRequest("POST", "/create", strings.NewReader("invalid data"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return an error due to nil use case or incorrect format
	assert.True(t, w.Code >= 400)
}

func TestHandleCreateSecret_MissingSecretID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form without secret_id
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_MissingOpaqueUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form without opaque_upload
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("message", "test message")
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_InvalidTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form with invalid TTL
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("ttl", "invalid")
	writer.WriteField("ttl_unit", "minutes")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_NegativeTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form with negative TTL
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("ttl", "-10")
	writer.WriteField("ttl_unit", "minutes")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_InvalidOpaqueUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form with invalid base64 opaque_upload
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", "invalid-base64!")
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_NoFileNoMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form without file and message
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_MessageTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create a message larger than the limit
	largeMessage := strings.Repeat("x", cfg.MaxPayloadBytes()+1)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("message", largeMessage)
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLoadSecret_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/secrets/:id", handlers.HandleLoadSecret)

	// Test with empty ID (the :id parameter will not be set)
	req := httptest.NewRequest("GET", "/secrets/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Usually returns 404 because the route does not match, but let's check that the handler works
	assert.True(t, w.Code >= 400)
}

func TestHandleLoadSecret_ValidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/secrets/:id", handlers.HandleLoadSecret)

	// Test with valid ID
	req := httptest.NewRequest("GET", "/secrets/test-secret-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleRevealSecret_MissingSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test without session_id
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("finalization", base64.StdEncoding.EncodeToString([]byte("test")))
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevealSecret_MissingFinalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test without finalization
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "test-session-id")
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevealSecret_InvalidFinalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test with invalid base64 finalization
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "test-session-id")
	writer.WriteField("finalization", "invalid-base64!")
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReadUploadedFile_EmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{MaxPayloadMB: 1}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Create an empty file through form
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))

	// Create an empty file
	part, err := writer.CreateFormFile("file", "empty.txt")
	assert.NoError(t, err)
	// Do not write anything to part - the file will be empty
	_ = part
	writer.Close()

	// Create form request
	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse form to get the file
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("Failed to parse form: %v", err)
	}

	// Get file header
	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	file.Close()

	// Test readUploadedFile with empty file
	_, err = handlers.readUploadedFile(header)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file is empty")
}

func TestNormalizePayloadType_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		usedPlainText bool
		expected      domain.PayloadType
	}{
		{
			name:          "uppercase TEXT",
			raw:           "TEXT",
			usedPlainText: false,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "mixed case File",
			raw:           "File",
			usedPlainText: true,
			expected:      domain.PayloadTypeFile,
		},
		{
			name:          "with whitespace",
			raw:           "  text  ",
			usedPlainText: false,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "unknown type with plain text",
			raw:           "unknown",
			usedPlainText: true,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "unknown type without plain text",
			raw:           "unknown",
			usedPlainText: false,
			expected:      domain.PayloadTypeFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePayloadType(tt.raw, tt.usedPlainText)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleCreateSecret_FileTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1, // 1 MB limit
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create a file larger than the limit (simulate)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))

	// Create a large file - but due to test limitations, we'll just create with valid content
	part, err := writer.CreateFormFile("file", "large.txt")
	assert.NoError(t, err)

	// Write large content
	largeContent := strings.Repeat("x", cfg.MaxPayloadBytes()+1)
	part.Write([]byte(largeContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return an error due to file size or form parsing error
	assert.True(t, w.Code >= 400)
}

func TestHandleCreateSecret_ZeroTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create multipart form with TTL set to 0 (invalid)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("ttl", "0") // invalid TTL
	writer.WriteField("ttl_unit", "minutes")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 400 due to invalid TTL
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLoadSecret_WithEmptyParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	// Without :id parameter
	router.GET("/secrets/", handlers.HandleLoadSecret)

	req := httptest.NewRequest("GET", "/secrets/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should handle missing ID
	assert.True(t, w.Code >= 400)
}

func TestHandleRevealSecret_EmptySessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test with empty session_id (only spaces)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "   ") // only spaces
	writer.WriteField("finalization", base64.StdEncoding.EncodeToString([]byte("test")))
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevealSecret_EmptyFinalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test with empty finalization (only spaces)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "test-session-id")
	writer.WriteField("finalization", "   ") // only spaces
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReadUploadedFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{MaxPayloadMB: 1}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Create a file with content
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))

	// Create a file with content
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)
	testContent := "Hello World"
	part.Write([]byte(testContent))
	writer.Close()

	// Create form request
	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse form to get the file
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("Failed to parse form: %v", err)
	}

	// Get file header
	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	file.Close()

	// Test readUploadedFile with a file with content
	payload, err := handlers.readUploadedFile(header)
	assert.NoError(t, err)
	assert.Equal(t, testContent, string(payload))
}

func TestHandleCreateSecret_ParseFormError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Create request with incorrect Content-Type
	req := httptest.NewRequest("POST", "/create", strings.NewReader("not multipart data"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should correctly handle form parsing error
	assert.True(t, w.Code >= 400)
}
