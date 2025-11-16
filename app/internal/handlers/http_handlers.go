package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/config"
	"github.com/gonfff/safex/app/internal/domain"
	"github.com/gonfff/safex/app/internal/usecases"
	"github.com/gonfff/safex/app/web"
)

// HTTPHandlers contains all HTTP handlers
type HTTPHandlers struct {
	cfg            config.Config
	createSecretUC *usecases.CreateSecretUseCase
	loadSecretUC   *usecases.LoadSecretUseCase
	deleteSecretUC *usecases.DeleteSecretUseCase
	opaqueAuthUC   *usecases.OpaqueAuthUseCase
	logger         zerolog.Logger
	templates      *template.Template
}

// NewHTTPHandlers creates a new instance of HTTP handlers
func NewHTTPHandlers(
	cfg config.Config,
	createSecretUC *usecases.CreateSecretUseCase,
	loadSecretUC *usecases.LoadSecretUseCase,
	deleteSecretUC *usecases.DeleteSecretUseCase,
	opaqueAuthUC *usecases.OpaqueAuthUseCase,
	logger zerolog.Logger,
) (*HTTPHandlers, error) {
	// Регистрируем кастомные валидаторы
	RegisterCustomValidators()

	tpl, err := web.Templates()
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	return &HTTPHandlers{
		cfg:            cfg,
		createSecretUC: createSecretUC,
		loadSecretUC:   loadSecretUC,
		deleteSecretUC: deleteSecretUC,
		opaqueAuthUC:   opaqueAuthUC,
		logger:         logger,
		templates:      tpl,
	}, nil
}

var errInvalidPinOrMissing = errors.New("file not found or invalid PIN")

// HandleHealth handles health check requests
func (h *HTTPHandlers) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleHome handles the home page
func (h *HTTPHandlers) HandleHome(c *gin.Context) {
	meta := h.buildPageMeta(c.Request, "", "Safex", "Safex - safe secrets exchange")
	data := homePageData{
		DefaultTTLMinutes:  int(h.cfg.DefaultTTL.Minutes()),
		MaxPayloadMB:       h.cfg.MaxPayloadMB,
		MaxPayloadBytes:    h.cfg.MaxPayloadBytes(),
		RateLimitPerMinute: h.cfg.RequestsPerMinute,
		Meta:               meta,
	}
	c.Status(http.StatusOK)
	h.renderTemplate(c, "home", data)
}

// HandleFAQ handles the FAQ page
func (h *HTTPHandlers) HandleFAQ(c *gin.Context) {
	meta := h.buildPageMeta(c.Request, "", "Safex", "Safex - safe secrets exchange")
	data := faqPageData{
		Meta: meta,
	}
	c.Status(http.StatusOK)
	h.renderTemplate(c, "faq", data)
}

// OpaqueRegisterStartResponse response to start OPAQUE registration
type OpaqueRegisterStartResponse struct {
	SecretID string `json:"secretId"`
	Response string `json:"response"`
}

// HandleOpaqueRegisterStart handles the start of OPAQUE registration
func (h *HTTPHandlers) HandleOpaqueRegisterStart(c *gin.Context) {
	var req OpaqueRegisterStartForm
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleValidationError(c, err)
		return
	}

	payload, err := base64.StdEncoding.DecodeString(req.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	secretID := uuid.New().String()
	response, err := h.opaqueAuthUC.StartRegistration(secretID, payload)
	if err != nil {
		h.logger.Error().Err(err).Msg("opaque registration response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "opaque registration failed"})
		return
	}

	c.JSON(http.StatusOK, OpaqueRegisterStartResponse{
		SecretID: secretID,
		Response: base64.StdEncoding.EncodeToString(response),
	})
}

// OpaqueLoginStartResponse response to start OPAQUE login
type OpaqueLoginStartResponse struct {
	SessionID string `json:"sessionId"`
	Response  string `json:"response"`
}

// HandleOpaqueLoginStart handles the start of OPAQUE login
func (h *HTTPHandlers) HandleOpaqueLoginStart(c *gin.Context) {
	var req OpaqueLoginStartForm
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleValidationError(c, err)
		return
	}

	payload, err := base64.StdEncoding.DecodeString(req.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		return
	}

	secret, err := h.loadSecretUC.GetMetadata(c.Request.Context(), req.SecretID)
	if err != nil {
		if errors.Is(err, domain.ErrSecretNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": errInvalidPinOrMissing.Error()})
			return
		}
		if errors.Is(err, domain.ErrSecretExpired) {
			c.JSON(http.StatusGone, gin.H{"error": "secret expired"})
			return
		}
		h.logger.Error().Err(err).Str("secret_id", req.SecretID).Msg("load metadata for opaque")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load secret"})
		return
	}

	if len(secret.OpaqueRecord) == 0 {
		h.logger.Error().Str("secret_id", req.SecretID).Msg("missing opaque record")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "opaque record missing"})
		return
	}

	sessionID, response, err := h.opaqueAuthUC.StartLogin(req.SecretID, secret.OpaqueRecord, payload)
	if err != nil {
		h.logger.Error().Err(err).Str("secret_id", req.SecretID).Msg("opaque login start")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "opaque login failed"})
		return
	}

	c.JSON(http.StatusOK, OpaqueLoginStartResponse{
		SessionID: sessionID,
		Response:  base64.StdEncoding.EncodeToString(response),
	})
}
