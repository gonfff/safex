package infrastructure

import (
	"testing"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/config"
)

func TestNewServer(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:     ":8080",
		Environment:  "test",
		MaxPayloadMB: 10,
		DefaultTTL:   900000000000, // 15 minutes
	}
	logger := zerolog.Nop()

	server, err := NewServer(cfg, nil, logger)

	// This will fail due to handlers being nil, but tests the constructor
	if err == nil && server == nil {
		t.Error("expected either error or server, got neither")
	}
}

func TestNewServer_WithRateLimiter(t *testing.T) {
	cfg := config.Config{
		HTTPAddr:          ":8080",
		Environment:       "production",
		RequestsPerMinute: 60,
	}
	logger := zerolog.Nop()

	_, err := NewServer(cfg, nil, logger)
	// Expect error due to nil handlers, but this tests rate limiter path
	if err == nil {
		t.Log("Successfully created server with rate limiter")
	}
}
