package server

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/backend/internal/config"
	"github.com/gonfff/safex/backend/internal/secret"
	"github.com/gonfff/safex/backend/internal/storage/blob"
	"github.com/gonfff/safex/backend/internal/storage/metadata"
)

func TestServerHealthz(t *testing.T) {
	srv := newTestServer(t, defaultTestConfig())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	srv.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if body := rec.Body.String(); !strings.Contains(body, "ok") {
		t.Fatalf("unexpected body %q", body)
	}
}

func TestCreateAndRevealSecretFlow(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.MaxPayloadMB = 1
	srv := newTestServer(t, cfg)

	payload := []byte("secret payload")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "secret.bin")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := writer.WriteField("pin", "1234"); err != nil {
		t.Fatalf("write pin: %v", err)
	}
	if err := writer.WriteField("ttl_minutes", "5"); err != nil {
		t.Fatalf("write ttl: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/secrets", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("HX-Request", "true")

	rec := httptest.NewRecorder()
	srv.engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	id := extractSecretID(t, rec.Body.String())
	if id == "" {
		t.Fatalf("secret id not found in response %q", rec.Body.String())
	}

	form := url.Values{}
	form.Set("secret_id", id)
	form.Set("pin", "1234")

	revealReq := httptest.NewRequest(http.MethodPost, "/secrets/reveal", strings.NewReader(form.Encode()))
	revealReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	revealReq.Header.Set("HX-Request", "true")

	revealRec := httptest.NewRecorder()
	srv.engine.ServeHTTP(revealRec, revealReq)

	if revealRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", revealRec.Code, revealRec.Body.String())
	}

	expectedPayload := base64.StdEncoding.EncodeToString(payload)
	if !strings.Contains(revealRec.Body.String(), expectedPayload) {
		t.Fatalf("payload not found in response: %s", revealRec.Body.String())
	}
}

func TestRateLimiterBlocksExcessRequests(t *testing.T) {
	cfg := defaultTestConfig()
	cfg.RequestsPerMinute = 1
	srv := newTestServer(t, cfg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request expected 200, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	srv.engine.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec2.Code)
	}
}

func newTestServer(t *testing.T, cfg config.Config) *Server {
	t.Helper()
	tempDir := t.TempDir()

	blobStore, err := blob.NewLocal(filepath.Join(tempDir, "blobs"))
	if err != nil {
		t.Fatalf("init blob store: %v", err)
	}
	metaStore, err := metadata.NewBolt(filepath.Join(tempDir, "metadata.db"))
	if err != nil {
		t.Fatalf("init metadata store: %v", err)
	}
	t.Cleanup(func() {
		_ = metaStore.Close()
	})

	if cfg.DefaultTTL == 0 {
		cfg.DefaultTTL = time.Minute
	}

	logger := zerolog.New(io.Discard)
	svc := secret.NewService(blobStore, metaStore, logger)
	srv, err := New(cfg, svc, logger)
	if err != nil {
		t.Fatalf("init server: %v", err)
	}
	return srv
}

func defaultTestConfig() config.Config {
	return config.Config{
		HTTPAddr:          ":0",
		MaxPayloadMB:      5,
		RequestsPerMinute: 10,
		DefaultTTL:        time.Minute,
		Environment:       "test",
	}
}

func extractSecretID(t *testing.T, body string) string {
	t.Helper()
	re := regexp.MustCompile(`data-secret-id="([^"]+)"`)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
