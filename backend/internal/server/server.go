package server

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/backend/internal/config"
	"github.com/gonfff/safex/backend/internal/secret"
	"github.com/gonfff/safex/backend/internal/storage/metadata"
	"github.com/gonfff/safex/backend/web"
)

// Server exposes HTTP endpoints backed by the secret service.
type Server struct {
	cfg       config.Config
	svc       *secret.Service
	logger    zerolog.Logger
	engine    *gin.Engine
	templates *template.Template
	httpSrv   *http.Server
}

// New wires up a Gin server instance with templates, static assets, and middleware.
func New(cfg config.Config, svc *secret.Service, logger zerolog.Logger) (*Server, error) {
	tpl, err := web.Templates()
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	staticFS, err := web.Static()
	if err != nil {
		return nil, fmt.Errorf("load static assets: %w", err)
	}

	switch strings.ToLower(cfg.Environment) {
	case "development", "dev":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery(), zerologMiddleware(logger))
	engine.StaticFS("/static", http.FS(staticFS))

	var limiter *rateLimiter
	if cfg.RequestsPerMinute > 0 {
		limiter = newRateLimiter(cfg.RequestsPerMinute, time.Minute)
		engine.Use(rateLimitMiddleware(limiter, logger))
	}

	s := &Server{
		cfg:       cfg,
		svc:       svc,
		logger:    logger,
		engine:    engine,
		templates: tpl,
	}
	s.registerRoutes()
	return s, nil
}

// Start begins serving HTTP requests until the context is canceled.
func (s *Server) Start(ctx context.Context) error {
	s.httpSrv = &http.Server{
		Addr:    s.cfg.HTTPAddr,
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	s.logger.Info().Str("addr", s.cfg.HTTPAddr).Msg("server started")

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpSrv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown server: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) registerRoutes() {
	s.engine.GET("/healthz", s.handleHealth)
	s.engine.GET("/", s.handleHome)
	s.engine.GET("/faq", s.handleFAQ)
	s.engine.POST("/secrets", s.handleCreateSecret)
	s.engine.GET("/secret/:id", s.handleLoadSecret)
	s.engine.GET("/secrets/:id", s.handleLoadSecret)
	s.engine.POST("/secrets/reveal", s.handleRevealSecret)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) handleHome(c *gin.Context) {
	meta := s.buildPageMeta(c.Request, "", "Safex", "Safex - safe secret exchange")
	data := homePageData{
		DefaultTTLMinutes:  int(s.cfg.DefaultTTL.Minutes()),
		MaxPayloadMB:       s.cfg.MaxPayloadMB,
		RateLimitPerMinute: s.cfg.RequestsPerMinute,
		Meta:               meta,
	}
	c.Status(http.StatusOK)
	s.renderTemplate(c, "home", data)
}

func (s *Server) handleFAQ(c *gin.Context) {
	meta := s.buildPageMeta(c.Request, "", "Safex", "Safex - safe secret exchange")
	data := faqPageData{
		Meta: meta,
	}
	c.Status(http.StatusOK)
	s.renderTemplate(c, "faq", data)
}

func (s *Server) handleCreateSecret(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(int64(s.cfg.MaxPayloadBytes()) + 1024); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		s.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	ttl := s.cfg.DefaultTTL
	if ttlStr := strings.TrimSpace(c.PostForm("ttl_minutes")); ttlStr != "" {
		minutes, err := strconv.Atoi(ttlStr)
		if err != nil || minutes <= 0 {
			s.renderCreateResult(c, http.StatusBadRequest, errors.New("TTL must be a positive number of minutes"))
			return
		}
		ttl = time.Duration(minutes) * time.Minute
	}

	message := strings.TrimSpace(c.PostForm("message"))
	payloadTypeRaw := strings.TrimSpace(strings.ToLower(c.PostForm("payload_type")))
	fileHeader, err := c.FormFile("file")
	var payload []byte
	input := secret.CreateInput{TTL: ttl}
	usedPlainText := false

	switch {
	case err == nil:
		payload, err = s.readUploadedFile(fileHeader)
		if err != nil {
			s.renderCreateResult(c, http.StatusBadRequest, err)
			return
		}
		input.FileName = fileHeader.Filename
		if ct := fileHeader.Header.Get("Content-Type"); ct != "" {
			input.ContentType = ct
		} else {
			input.ContentType = http.DetectContentType(payload)
		}
	case errors.Is(err, http.ErrMissingFile):
		if message == "" {
			s.renderCreateResult(c, http.StatusBadRequest, errors.New("File or message is required"))
			return
		}
		usedPlainText = true
		payload = []byte(message)
		if len(payload) > s.cfg.MaxPayloadBytes() {
			s.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("message exceeds %d bytes", s.cfg.MaxPayloadBytes()))
			return
		}
		input.FileName = "message.txt"
		input.ContentType = "text/plain; charset=utf-8"
	default:
		s.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("failed to read file: %w", err))
		return
	}
	input.Payload = payload
	input.PayloadType = normalizePayloadType(payloadTypeRaw, usedPlainText)
	ctx := c.Request.Context()
	rec, err := s.svc.Create(ctx, input)
	if err != nil {
		s.logger.Error().Err(err).Msg("create secret")
		s.renderCreateResult(c, http.StatusInternalServerError, errors.New("failed to save secret"))
		return
	}

	data := createResultData{
		Record:    rec,
		TTL:       ttl,
		MaxBytes:  s.cfg.MaxPayloadBytes(),
		SharePath: fmt.Sprintf("/secret/%s", rec.ID),
		ShareURL:  s.makeShareURL(c.Request, rec.ID),
	}
	c.Status(http.StatusCreated)
	s.renderTemplate(c, "createResult", data)
}

func (s *Server) handleLoadSecret(c *gin.Context) {
	id := c.Param("id")
	metaPath := c.Request.URL.Path
	if id != "" {
		metaPath = fmt.Sprintf("/secret/%s", id)
	}
	meta := s.buildPageMeta(c.Request, metaPath, "Safex ", "Safex — decrypt your secret message via the link")
	if id == "" {
		c.Status(http.StatusBadRequest)
		s.renderTemplate(c, "retrieve", getSecretData{
			Error: errors.New("secret ID is required"),
			Meta:  meta,
		})
		return
	}

	data := getSecretData{
		SecretID: id,
		Title:    "Safex",
		Meta:     meta,
	}
	c.Status(http.StatusOK)
	s.renderTemplate(c, "retrieve", data)
}

func (s *Server) handleRevealSecret(c *gin.Context) {
	id := strings.TrimSpace(c.PostForm("secret_id"))
	if id == "" {
		s.renderRevealResult(c, http.StatusBadRequest, errors.New("secret ID is required"), nil, nil)
		return
	}

	rec, payload, err := s.svc.Load(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, metadata.ErrNotFound) {
			s.renderRevealResult(c, http.StatusNotFound, errors.New("secret not found"), nil, nil)
			return
		}
		if strings.Contains(err.Error(), "expired") {
			s.renderRevealResult(c, http.StatusGone, errors.New("secret expired"), nil, nil)
			return
		}
		s.logger.Error().Err(err).Str("secret_id", id).Msg("load secret")
		s.renderRevealResult(c, http.StatusInternalServerError, errors.New("failed to load secret"), nil, nil)
		return
	}

	s.renderRevealResult(c, http.StatusOK, nil, &rec, payload)
}

func (s *Server) renderCreateResult(c *gin.Context, status int, renderErr error) {
	c.Status(status)
	s.renderTemplate(c, "createResult", createResultData{Error: renderErr})
}

func (s *Server) renderRevealResult(c *gin.Context, status int, renderErr error, rec *metadata.MetadataRecord, payload []byte) {
	c.Status(status)
	data := revealResultData{
		Error: renderErr,
	}
	if rec != nil {
		data.Record = rec
		if rec.PayloadType == metadata.PayloadTypeText {
			data.IsText = true
		}
	}
	if len(payload) > 0 {
		data.PayloadBase64 = base64.StdEncoding.EncodeToString(payload)
		if rec != nil && strings.HasPrefix(rec.ContentType, "text/") && utf8.Valid(payload) {
			data.PayloadText = string(payload)
			data.IsText = true
		}
	}
	s.renderTemplate(c, "revealResult", data)
}

func (s *Server) readUploadedFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer file.Close()

	limit := int64(s.cfg.MaxPayloadBytes())
	payload, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if int64(len(payload)) > limit {
		return nil, fmt.Errorf("file size exceeds %d bytes", limit)
	}
	if len(payload) == 0 {
		return nil, errors.New("file is empty")
	}
	return payload, nil
}

func (s *Server) renderTemplate(c *gin.Context, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		s.logger.Error().Err(err).Str("template", name).Msg("render template")
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

type homePageData struct {
	DefaultTTLMinutes  int
	MaxPayloadMB       int
	RateLimitPerMinute int
	Meta               pageMeta
}

type faqPageData struct {
	Meta pageMeta
}

type createResultData struct {
	Error     error
	Record    metadata.MetadataRecord
	TTL       time.Duration
	MaxBytes  int
	SharePath string
	ShareURL  string
}

type revealResultData struct {
	Error         error
	Record        *metadata.MetadataRecord
	PayloadBase64 string
	PayloadText   string
	IsText        bool
}

type getSecretData struct {
	Title    string
	SecretID string
	Error    error
	Meta     pageMeta
}

func (s *Server) makeShareURL(r *http.Request, id string) string {
	return s.makeAbsoluteURL(r, fmt.Sprintf("/secret/%s", id))
}

type pageMeta struct {
	Canonical     string
	OGTitle       string
	OGDescription string
	OGType        string
	OGImage       string
}

func (s *Server) buildPageMeta(r *http.Request, path, title, description string) pageMeta {
	if path == "" && r.URL != nil {
		path = r.URL.Path
	}
	if path == "" {
		path = "/"
	}
	return pageMeta{
		Canonical:     s.makeAbsoluteURL(r, path),
		OGTitle:       title,
		OGDescription: description,
		OGType:        "website",
	}
}

func (s *Server) makeAbsoluteURL(r *http.Request, path string) string {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	scheme := "https"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		parts := strings.Split(proto, ",")
		s := strings.TrimSpace(parts[0])
		if s != "" {
			scheme = s
		}
	} else if r.TLS == nil {
		scheme = "http"
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host != "" {
		parts := strings.Split(host, ",")
		host = strings.TrimSpace(parts[0])
	}
	if host == "" {
		host = r.Host
	}
	if host == "" {
		addr := s.cfg.HTTPAddr
		if strings.HasPrefix(addr, ":") {
			host = "localhost" + addr
		} else {
			host = addr
		}
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}

func normalizePayloadType(raw string, usedPlainText bool) metadata.PayloadType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(metadata.PayloadTypeText):
		return metadata.PayloadTypeText
	case string(metadata.PayloadTypeFile):
		return metadata.PayloadTypeFile
	}
	if usedPlainText {
		return metadata.PayloadTypeText
	}
	return metadata.PayloadTypeFile
}
