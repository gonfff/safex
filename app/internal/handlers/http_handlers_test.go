package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/config"
)

func TestHTTPHandlers_HandleHealth(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("failed to create handlers: %v", err)
	}

	router := gin.New()
	router.GET("/health", handlers.HandleHealth)

	// Test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	expectedBody := `{"status":"ok"}`
	if w.Body.String() != expectedBody {
		t.Errorf("expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestHTTPHandlers_HandleHome(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		HTTPAddr:          ":8080",
		MaxPayloadMB:      10,
		RequestsPerMinute: 10,
		DefaultTTL:        900000000000, // 15 minutes in nanoseconds
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("failed to create handlers: %v", err)
	}

	router := gin.New()
	router.GET("/", handlers.HandleHome)

	// Test request
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should contain HTML content
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected HTML content type, got %s", w.Header().Get("Content-Type"))
	}
}

func TestHTTPHandlers_HandleFAQ(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	cfg := config.Config{HTTPAddr: ":8080"}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("failed to create handlers: %v", err)
	}

	router := gin.New()
	router.GET("/faq", handlers.HandleFAQ)

	// Test request
	req := httptest.NewRequest("GET", "/faq", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Should contain HTML content
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected HTML content type, got %s", w.Header().Get("Content-Type"))
	}
}

// TestDecodeBase64Field убран - декодирование base64 теперь в валидаторах

func TestNewHTTPHandlers_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{}
	logger := zerolog.Nop()

	// Create handlers - should work without error normally
	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if handlers == nil {
		t.Fatal("Expected handlers to be created")
	}
}
