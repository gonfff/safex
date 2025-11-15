package handlers

import (
	"crypto/tls"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gonfff/safex/app/internal/config"
	"github.com/gonfff/safex/app/internal/domain"
	"github.com/rs/zerolog"
)

func TestHTTPHandlers_buildPageMeta(t *testing.T) {
	cfg := config.Config{HTTPAddr: ":8080"}
	h := &HTTPHandlers{cfg: cfg}

	tests := []struct {
		name        string
		path        string
		title       string
		description string
		requestURL  string
		want        pageMeta
	}{
		{
			name:        "with provided path",
			path:        "/test",
			title:       "Test Title",
			description: "Test Description",
			requestURL:  "http://localhost:8080/other",
			want: pageMeta{
				Canonical:     "http://localhost:8080/test",
				OGTitle:       "Test Title",
				OGDescription: "Test Description",
				OGType:        "website",
			},
		},
		{
			name:        "with empty path uses request URL",
			path:        "",
			title:       "Test Title",
			description: "Test Description",
			requestURL:  "http://localhost:8080/from-request",
			want: pageMeta{
				Canonical:     "http://localhost:8080/from-request",
				OGTitle:       "Test Title",
				OGDescription: "Test Description",
				OGType:        "website",
			},
		},
		{
			name:        "with empty everything defaults to root",
			path:        "",
			title:       "Test Title",
			description: "Test Description",
			requestURL:  "http://localhost:8080/",
			want: pageMeta{
				Canonical:     "http://localhost:8080/",
				OGTitle:       "Test Title",
				OGDescription: "Test Description",
				OGType:        "website",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestURL, nil)
			got := h.buildPageMeta(req, tt.path, tt.title, tt.description)

			if got.Canonical != tt.want.Canonical {
				t.Errorf("buildPageMeta() Canonical = %v, want %v", got.Canonical, tt.want.Canonical)
			}
			if got.OGTitle != tt.want.OGTitle {
				t.Errorf("buildPageMeta() OGTitle = %v, want %v", got.OGTitle, tt.want.OGTitle)
			}
			if got.OGDescription != tt.want.OGDescription {
				t.Errorf("buildPageMeta() OGDescription = %v, want %v", got.OGDescription, tt.want.OGDescription)
			}
			if got.OGType != tt.want.OGType {
				t.Errorf("buildPageMeta() OGType = %v, want %v", got.OGType, tt.want.OGType)
			}
		})
	}
}

func TestHTTPHandlers_makeShareURL(t *testing.T) {
	cfg := config.Config{HTTPAddr: ":8080"}
	h := &HTTPHandlers{cfg: cfg}

	tests := []struct {
		name       string
		id         string
		requestURL string
		want       string
	}{
		{
			name:       "basic share URL",
			id:         "test-id",
			requestURL: "http://localhost:8080/",
			want:       "http://localhost:8080/secrets/test-id",
		},
		{
			name:       "HTTPS request",
			id:         "secure-id",
			requestURL: "https://example.com/",
			want:       "https://example.com/secrets/secure-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestURL, nil)
			got := h.makeShareURL(req, tt.id)

			if got != tt.want {
				t.Errorf("makeShareURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPHandlers_makeAbsoluteURL(t *testing.T) {
	cfg := config.Config{HTTPAddr: ":8080"}
	h := &HTTPHandlers{cfg: cfg}

	tests := []struct {
		name       string
		path       string
		requestURL string
		headers    map[string]string
		useTLS     bool
		want       string
	}{
		{
			name:       "basic HTTP request",
			path:       "/test",
			requestURL: "http://localhost:8080/",
			want:       "http://localhost:8080/test",
		},
		{
			name:       "HTTPS request",
			path:       "/test",
			requestURL: "https://localhost:8080/",
			useTLS:     true,
			want:       "https://localhost:8080/test",
		},
		{
			name:       "with X-Forwarded-Proto header",
			path:       "/test",
			requestURL: "http://localhost:8080/",
			headers:    map[string]string{"X-Forwarded-Proto": "https"},
			want:       "https://localhost:8080/test",
		},
		{
			name:       "with X-Forwarded-Host header",
			path:       "/test",
			requestURL: "http://localhost:8080/",
			headers:    map[string]string{"X-Forwarded-Host": "example.com"},
			want:       "http://example.com/test",
		},
		{
			name:       "path without leading slash",
			path:       "test",
			requestURL: "http://localhost:8080/",
			want:       "http://localhost:8080/test",
		},
		{
			name:       "already absolute URL",
			path:       "https://external.com/test",
			requestURL: "http://localhost:8080/",
			want:       "https://external.com/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestURL, nil)

			// Set TLS if needed
			if tt.useTLS {
				req.TLS = &tls.ConnectionState{}
			}

			// Set headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := h.makeAbsoluteURL(req, tt.path)

			if got != tt.want {
				t.Errorf("makeAbsoluteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateResultData(t *testing.T) {
	secret := domain.Secret{
		ID:          "test-id",
		FileName:    "test.txt",
		ContentType: "text/plain",
		Size:        100,
		ExpiresAt:   time.Now().Add(time.Hour),
		PayloadType: domain.PayloadTypeText,
	}

	data := createResultData{
		Record:    secret,
		TTL:       time.Hour,
		MaxBytes:  1024,
		SharePath: "/secrets/test-id",
		ShareURL:  "https://example.com/secrets/test-id",
	}

	if data.Record.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", data.Record.ID)
	}
	if data.TTL != time.Hour {
		t.Errorf("Expected TTL to be 1 hour, got %v", data.TTL)
	}
}

func TestRevealResultData(t *testing.T) {
	secret := &domain.Secret{
		ID:          "test-id",
		FileName:    "test.txt",
		ContentType: "text/plain",
		Size:        100,
		ExpiresAt:   time.Now().Add(time.Hour),
		PayloadType: domain.PayloadTypeText,
	}

	payload := []byte("test payload")
	payloadBase64 := base64.StdEncoding.EncodeToString(payload)

	data := revealResultData{
		Record:        secret,
		PayloadBase64: payloadBase64,
		PayloadText:   string(payload),
		IsText:        true,
	}

	if data.Record.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", data.Record.ID)
	}
	if data.PayloadText != "test payload" {
		t.Errorf("Expected PayloadText to be 'test payload', got %s", data.PayloadText)
	}
	if !data.IsText {
		t.Error("Expected IsText to be true")
	}
}

func TestNormalizePayloadTypeInternal(t *testing.T) {
	// Test как функция пакета
	tests := []struct {
		name          string
		raw           string
		usedPlainText bool
		expected      domain.PayloadType
	}{
		{
			name:          "text type",
			raw:           "text",
			usedPlainText: true,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "file type",
			raw:           "file",
			usedPlainText: false,
			expected:      domain.PayloadTypeFile,
		},
		{
			name:          "auto detect text",
			raw:           "",
			usedPlainText: true,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "auto detect file",
			raw:           "",
			usedPlainText: false,
			expected:      domain.PayloadTypeFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePayloadType(tt.raw, tt.usedPlainText)
			if result != tt.expected {
				t.Errorf("normalizePayloadType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHTTPHandlers_renderTemplate_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{HTTPAddr: ":8080"}
	logger := zerolog.Nop()

	// Create handlers with proper initialization but will still fail on non-existent template
	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	// Create a gin context with test response writer
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// renderTemplate with non-existent template should handle error gracefully
	handlers.renderTemplate(c, "nonexistent", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHTTPHandlers_renderRevealResult_WithInvalidPin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{HTTPAddr: ":8080"}
	logger := zerolog.Nop()

	h, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with invalid PIN error
	h.renderRevealResult(c, http.StatusBadRequest, errInvalidPinOrMissing, nil, nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHTTPHandlers_renderRevealResult_WithSecretAndPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{HTTPAddr: ":8080"}
	logger := zerolog.Nop()

	h, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	secret := &domain.Secret{
		ID:          "test-id",
		FileName:    "test.txt",
		ContentType: "text/plain; charset=utf-8",
		PayloadType: domain.PayloadTypeText,
	}
	payload := []byte("Hello World")

	h.renderRevealResult(c, http.StatusOK, nil, secret, payload)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHTTPHandlers_renderRevealResult_WithBinaryPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{HTTPAddr: ":8080"}
	logger := zerolog.Nop()

	h, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	secret := &domain.Secret{
		ID:          "test-id",
		FileName:    "test.bin",
		ContentType: "application/octet-stream",
		PayloadType: domain.PayloadTypeFile,
	}
	payload := []byte{0x00, 0x01, 0x02, 0xFF}

	h.renderRevealResult(c, http.StatusOK, nil, secret, payload)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHTTPHandlers_makeAbsoluteURL_EdgeCases(t *testing.T) {
	cfg := config.Config{HTTPAddr: "example.com:9000"}
	h := &HTTPHandlers{cfg: cfg}

	tests := []struct {
		name       string
		path       string
		requestURL string
		headers    map[string]string
		host       string
		want       string
	}{
		{
			name:       "multiple X-Forwarded-Proto values",
			path:       "/test",
			requestURL: "http://localhost:8080/",
			headers:    map[string]string{"X-Forwarded-Proto": "https,http"},
			want:       "https://localhost:8080/test",
		},
		{
			name:       "multiple X-Forwarded-Host values",
			path:       "/test",
			requestURL: "http://localhost:8080/",
			headers:    map[string]string{"X-Forwarded-Host": "proxy1.com,proxy2.com"},
			want:       "http://proxy1.com/test",
		},
		{
			name:       "empty host uses config address",
			path:       "/test",
			requestURL: "http:///",
			want:       "http://example.com:9000/test",
		},
		{
			name:       "config address starts with colon",
			path:       "/test",
			requestURL: "http:///",
			want:       "http://localhost:9000/test",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if i == 3 { // Last test case - change config
				h.cfg.HTTPAddr = ":9000"
			}

			req := httptest.NewRequest("GET", tt.requestURL, nil)

			// Remove host if needed for empty host test
			if tt.host == "" && tt.requestURL == "http:///" {
				req.Host = ""
			}

			// Set headers
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := h.makeAbsoluteURL(req, tt.path)

			if got != tt.want {
				t.Errorf("makeAbsoluteURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPHandlers_buildPageMeta_EmptyPath(t *testing.T) {
	cfg := config.Config{HTTPAddr: ":8080"}
	h := &HTTPHandlers{cfg: cfg}

	req := httptest.NewRequest("GET", "http://localhost:8080/", nil)

	// Test with empty path and empty request URL path
	req.URL.Path = ""
	got := h.buildPageMeta(req, "", "Title", "Description")

	// Should default to "/"
	if got.Canonical != "http://localhost:8080/" {
		t.Errorf("buildPageMeta() with empty path, got Canonical = %v", got.Canonical)
	}
	if got.OGTitle != "Title" {
		t.Errorf("buildPageMeta() OGTitle = %v, want Title", got.OGTitle)
	}
}
