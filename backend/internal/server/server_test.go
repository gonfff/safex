package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
	"github.com/gonfff/safex/backend/internal/opaqueauth"
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

	testClient := opaqueauth.NewClient()
	payload := []byte("secret payload")
	password := "123456"

	regHandle, regMessage, err := testClient.StartRegistration(password)
	if err != nil {
		t.Fatalf("start registration: %v", err)
	}
	regReq := httptest.NewRequest(http.MethodPost, "/opaque/register/start", bytes.NewReader(mustJSON(map[string]string{
		"request": base64.StdEncoding.EncodeToString(regMessage),
	})))
	regReq.Header.Set("Content-Type", "application/json")

	regRec := httptest.NewRecorder()
	srv.engine.ServeHTTP(regRec, regReq)
	if regRec.Code != http.StatusOK {
		t.Fatalf("register start failed: %d %s", regRec.Code, regRec.Body.String())
	}

	var regResp opaqueRegisterStartResponse
	if err := json.Unmarshal(regRec.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	respBytes, err := base64.StdEncoding.DecodeString(regResp.Response)
	if err != nil {
		t.Fatalf("decode server register message: %v", err)
	}
	upload, _, err := testClient.FinishRegistration(regHandle, password, respBytes)
	if err != nil {
		t.Fatalf("finish registration: %v", err)
	}
	uploadB64 := base64.StdEncoding.EncodeToString(upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "secret.bin")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := writer.WriteField("ttl_minutes", "5"); err != nil {
		t.Fatalf("write ttl: %v", err)
	}
	if err := writer.WriteField("secret_id", regResp.SecretID); err != nil {
		t.Fatalf("write secret id: %v", err)
	}
	if err := writer.WriteField("opaque_upload", uploadB64); err != nil {
		t.Fatalf("write opaque upload: %v", err)
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

	loginHandle, loginMessage, err := testClient.StartLogin(password)
	if err != nil {
		t.Fatalf("start login: %v", err)
	}
	loginReq := httptest.NewRequest(http.MethodPost, "/opaque/login/start", bytes.NewReader(mustJSON(map[string]string{
		"secretId": id,
		"request":  base64.StdEncoding.EncodeToString(loginMessage),
	})))
	loginReq.Header.Set("Content-Type", "application/json")

	loginRec := httptest.NewRecorder()
	srv.engine.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusOK {
		t.Fatalf("login start failed: %d %s", loginRec.Code, loginRec.Body.String())
	}

	var loginResp opaqueLoginStartResponse
	if err := json.Unmarshal(loginRec.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	ke2Bytes, err := base64.StdEncoding.DecodeString(loginResp.Response)
	if err != nil {
		t.Fatalf("decode login response payload: %v", err)
	}
	ke3, _, _, err := testClient.FinishLogin(loginHandle, password, ke2Bytes)
	if err != nil {
		t.Fatalf("finish login: %v", err)
	}

	form := url.Values{}
	form.Set("secret_id", id)
	form.Set("session_id", loginResp.SessionID)
	form.Set("finalization", base64.StdEncoding.EncodeToString(ke3))

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

func TestRevealSecretUnknownIDReturnsGenericError(t *testing.T) {
	srv := newTestServer(t, defaultTestConfig())

	form := url.Values{}
	form.Set("secret_id", "missing-id")
	form.Set("session_id", "missing-session")
	form.Set("finalization", base64.StdEncoding.EncodeToString([]byte("garbage")))

	req := httptest.NewRequest(http.MethodPost, "/secrets/reveal", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	rec := httptest.NewRecorder()
	srv.engine.ServeHTTP(rec, req)

	body := rec.Body.String()
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d body=%s", rec.Code, body)
	}
	if !strings.Contains(body, errInvalidPinOrMissing.Error()) {
		t.Fatalf("expected generic error message, got %q", body)
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
	if cfg.OpaqueSessionTTL == 0 {
		cfg.OpaqueSessionTTL = time.Minute
	}

	logger := zerolog.New(io.Discard)
	svc := secret.NewService(blobStore, metaStore, logger)
	opaqueMgr, err := opaqueauth.NewManager(cfg)
	if err != nil {
		t.Fatalf("init opaque manager: %v", err)
	}
	srv, err := New(cfg, svc, opaqueMgr, logger)
	if err != nil {
		t.Fatalf("init server: %v", err)
	}
	return srv
}

func defaultTestConfig() config.Config {
	privateKey := make([]byte, 32)
	seed := make([]byte, 64)
	rand.Read(privateKey)
	rand.Read(seed)
	return config.Config{
		HTTPAddr:          ":0",
		MaxPayloadMB:      5,
		RequestsPerMinute: 10,
		DefaultTTL:        time.Minute,
		Environment:       "test",
		OpaqueServerID:    "",
		OpaquePrivateKey:  base64.StdEncoding.EncodeToString(privateKey),
		OpaqueOPRFSeed:    base64.StdEncoding.EncodeToString(seed),
		OpaqueSessionTTL:  time.Minute,
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

func mustJSON(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
