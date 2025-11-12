package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/gonfff/safex/backend/internal/config"
	"github.com/gonfff/safex/backend/internal/secret"
	"github.com/gonfff/safex/backend/internal/storage/metadata"
	"github.com/gonfff/safex/backend/web"
)

// Server wires HTTP handlers with the secret service.
type Server struct {
	service   *secret.Service
	logger    zerolog.Logger
	cfg       config.Config
	httpSrv   *http.Server
	templates *template.Template
	staticFS  http.Handler
}

// viewData carries data passed into HTML templates.
type viewData struct {
	Title           string
	Active          string
	PrefillSecret   string
	ContentTemplate string
}

// New instantiates the HTTP server using Go's net/http stack.
func New(cfg config.Config, svc *secret.Service, logger zerolog.Logger) *Server {
	tmpl, err := web.Templates()
	if err != nil {
		panic(fmt.Sprintf("failed to parse templates: %v", err))
	}
	staticFS, err := web.Static()
	if err != nil {
		panic(fmt.Sprintf("failed to load static assets: %v", err))
	}

	mux := http.NewServeMux()
	srv := &Server{
		service:   svc,
		logger:    logger,
		cfg:       cfg,
		templates: tmpl,
		staticFS:  http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))),
	}

	mux.HandleFunc("/healthcheck", srv.handleHealth)
	mux.Handle("/static/", srv.staticFS)

	mux.HandleFunc("/", srv.pageCreate)
	mux.HandleFunc("/receive", srv.pageReceive)
	mux.HandleFunc("/docs", srv.pageDocs)

	mux.HandleFunc("/api/v1/secrets", srv.routeSecrets)
	mux.HandleFunc("/api/v1/secrets/", srv.routeSecretsWithID)

	srv.httpSrv = &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: srv.logRequests(mux),
	}
	return srv
}

// Start runs the HTTP server until context is canceled.
func (s *Server) Start(ctx context.Context) error {
	shutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(ctxShutdown)
		close(shutdown)
	}()
	s.logger.Info().Str("addr", s.httpSrv.Addr).Msg("listening")
	err := s.httpSrv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) && err != nil {
		return err
	}
	<-shutdown
	return nil
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info().Str("method", r.Method).Str("path", r.URL.Path).Dur("duration", time.Since(start)).Msg("request")
	})
}

// HTML pages -----------------------------------------------------------------

func (s *Server) pageCreate(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		s.notFound(w)
		return
	}
	data := viewData{
		Title:           "Safex — создать секрет",
		Active:          "create",
		ContentTemplate: "create-content",
	}
	s.render(w, data)
}

func (s *Server) pageReceive(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/receive" {
		s.notFound(w)
		return
	}
	prefill := r.URL.Query().Get("secret")
	data := viewData{
		Title:           "Safex — получить секрет",
		Active:          "receive",
		PrefillSecret:   prefill,
		ContentTemplate: "receive-content",
	}
	s.render(w, data)
}

func (s *Server) pageDocs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/docs" {
		s.notFound(w)
		return
	}
	data := viewData{
		Title:           "Safex — документация",
		Active:          "docs",
		ContentTemplate: "docs-content",
	}
	s.render(w, data)
}

func (s *Server) render(w http.ResponseWriter, data viewData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, "base", data); err != nil {
		s.logger.Error().Err(err).Msg("render template")
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

// API routes -----------------------------------------------------------------

func (s *Server) routeSecrets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.handleCreateSecret(w, r)
	default:
		s.methodNotAllowed(w)
	}
}

func (s *Server) routeSecretsWithID(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/v1/secrets/") {
		s.notFound(w)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/secrets/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		s.notFound(w)
		return
	}
	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			s.handleGetSecret(w, r, id)
		case http.MethodDelete:
			s.handleDeleteSecret(w, r, id)
		default:
			s.methodNotAllowed(w)
		}
		return
	}

	if len(parts) == 2 && parts[1] == "payload" {
		if r.Method == http.MethodGet {
			s.handleDownloadSecret(w, r, id)
			return
		}
		s.methodNotAllowed(w)
		return
	}

	s.notFound(w)
}

func (s *Server) handleCreateSecret(w http.ResponseWriter, r *http.Request) {
	req := r.Clone(r.Context())
	req.Body = http.MaxBytesReader(w, req.Body, s.maxPayload+1<<20)
	if err := req.ParseMultipartForm(s.maxPayload + 1<<20); err != nil {
		s.respondErr(w, http.StatusBadRequest, "invalid form: %v", err)
		return
	}
	file, header, err := req.FormFile("payload")
	if err != nil {
		s.respondErr(w, http.StatusBadRequest, "missing payload: %v", err)
		return
	}
	defer file.Close()
	payload, err := readAll(file, s.maxPayload)
	if err != nil {
		s.respondErr(w, http.StatusBadRequest, "%s", err.Error())
		return
	}

	filename := req.FormValue("filename")
	if filename == "" && header != nil {
		filename = header.Filename
	}
	contentType := req.FormValue("contentType")
	if contentType == "" && header != nil {
		contentType = header.Header.Get("Content-Type")
	}
	ttlSeconds := parseInt64(req.FormValue("ttlSeconds"), int64(s.defaultTTL.Seconds()))
	if ttlSeconds <= 0 {
		ttlSeconds = int64(s.defaultTTL.Seconds())
	}
	oneTime := req.FormValue("oneTime") == "true"

	record, err := s.service.Create(r.Context(), secret.CreateInput{
		FileName:    filename,
		ContentType: contentType,
		Payload:     payload,
		TTL:         time.Duration(ttlSeconds) * time.Second,
		OneTime:     oneTime,
	})
	if err != nil {
		s.respondErr(w, http.StatusBadRequest, "%s", err.Error())
		return
	}

	s.respondJSON(w, http.StatusCreated, map[string]any{
		"id":        record.ID,
		"expiresAt": record.ExpiresAt,
		"oneTime":   record.OneTime,
	})
}

func (s *Server) handleGetSecret(w http.ResponseWriter, r *http.Request, id string) {
	rec, err := s.service.Metadata(r.Context(), id)
	if err != nil {
		s.respondFromError(w, err)
		return
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"id":          rec.ID,
		"fileName":    rec.FileName,
		"contentType": rec.ContentType,
		"size":        rec.Size,
		"expiresAt":   rec.ExpiresAt,
		"oneTime":     rec.OneTime,
	})
}

func (s *Server) handleDownloadSecret(w http.ResponseWriter, r *http.Request, id string) {
	rec, payload, err := s.service.Fetch(r.Context(), id)
	if err != nil {
		s.respondFromError(w, err)
		return
	}
	contentType := "application/octet-stream"
	if rec.FileName != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", rec.FileName))
	}
	if rec.ContentType != "" {
		contentType = rec.ContentType
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(payload)
}

func (s *Server) handleDeleteSecret(w http.ResponseWriter, r *http.Request, id string) {
	if err := s.service.Delete(r.Context(), id); err != nil {
		s.respondFromError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) respondFromError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, metadata.ErrNotFound):
		s.respondErr(w, http.StatusNotFound, "secret not found")
	case errors.Is(err, metadata.ErrExpired):
		s.respondErr(w, http.StatusGone, "secret expired")
	default:
		s.logger.Error().Err(err).Msg("secret service error")
		s.respondErr(w, http.StatusInternalServerError, "%s", err.Error())
	}
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func (s *Server) respondErr(w http.ResponseWriter, status int, format string, args ...any) {
	s.respondJSON(w, status, map[string]string{"error": fmt.Sprintf(format, args...)})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) methodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) notFound(w http.ResponseWriter) {
	http.Error(w, "not found", http.StatusNotFound)
}

// Helpers --------------------------------------------------------------------

func readAll(file multipart.File, limit int64) ([]byte, error) {
	var buf bytes.Buffer
	n, err := io.CopyN(&buf, file, limit+1)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, err
	}
	if n > limit {
		return nil, fmt.Errorf("payload larger than %d bytes", limit)
	}
	return buf.Bytes(), nil
}

func parseInt64(val string, def int64) int64 {
	if val == "" {
		return def
	}
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return def
	}
	return v
}
