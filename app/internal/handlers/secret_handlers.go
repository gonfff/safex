package handlers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/gonfff/safex/app/internal/domain"
	"github.com/gonfff/safex/app/internal/opaqueauth"
	"github.com/gonfff/safex/app/internal/usecases"
)

// HandleCreateSecret handles secret creation
func (h *HTTPHandlers) HandleCreateSecret(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(int64(h.cfg.MaxPayloadBytes()) + 1024); err != nil && !errors.Is(err, http.ErrNotMultipart) {
		h.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	var form CreateSecretForm
	if err := c.ShouldBind(&form); err != nil {
		h.HandleValidationError(c, err)
		return
	}

	// Получаем TTL из формы или используем дефолтный
	ttl := form.GetTTLDuration(h.cfg.DefaultTTL)

	// Декодируем opaque upload
	opaqueUpload, err := base64.StdEncoding.DecodeString(form.OpaqueUpload)
	if err != nil {
		h.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("invalid opaque upload: %w", err))
		return
	}

	input := usecases.CreateSecretInput{
		ID:           form.SecretID,
		TTL:          ttl,
		OpaqueRecord: opaqueUpload,
	}

	var payload []byte
	usedPlainText := false

	fileHeader, fileErr := c.FormFile("file")
	switch {
	case fileErr == nil:
		payload, err = h.readUploadedFile(fileHeader)
		if err != nil {
			h.renderCreateResult(c, http.StatusBadRequest, err)
			return
		}
		input.FileName = fileHeader.Filename
		if ct := fileHeader.Header.Get("Content-Type"); ct != "" {
			input.ContentType = ct
		} else {
			input.ContentType = http.DetectContentType(payload)
		}
	case errors.Is(fileErr, http.ErrMissingFile):
		if form.Message == "" {
			h.renderCreateResult(c, http.StatusBadRequest, errors.New("file or message is required"))
			return
		}
		usedPlainText = true
		payload = []byte(form.Message)
		if len(payload) > h.cfg.MaxPayloadBytes() {
			h.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("message exceeds %d bytes", h.cfg.MaxPayloadBytes()))
			return
		}
		input.FileName = "message.txt"
		input.ContentType = "text/plain; charset=utf-8"
	default:
		h.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("failed to read file: %w", fileErr))
		return
	}

	input.Payload = payload
	input.PayloadType = normalizePayloadType(form.PayloadType, usedPlainText)

	ctx := c.Request.Context()
	secret, err := h.createSecretUC.Execute(ctx, input)
	if err != nil {
		h.logger.Error().Err(err).Msg("create secret")
		h.renderCreateResult(c, http.StatusInternalServerError, errors.New("failed to save secret"))
		return
	}

	data := createResultData{
		Record:    *secret,
		TTL:       ttl,
		MaxBytes:  h.cfg.MaxPayloadBytes(),
		SharePath: fmt.Sprintf("/secrets/%s", secret.ID),
		ShareURL:  h.makeShareURL(c.Request, secret.ID),
	}
	c.Status(http.StatusCreated)
	h.renderTemplate(c, "createResult", data)
}

// HandleLoadSecret handles loading the secret page
func (h *HTTPHandlers) HandleLoadSecret(c *gin.Context) {
	id := c.Param("id")
	metaPath := c.Request.URL.Path
	if id != "" {
		metaPath = fmt.Sprintf("/secrets/%s", id)
	}
	meta := h.buildPageMeta(c.Request, metaPath, "Safex ", "Safex — decrypt your secret message via the link")
	if id == "" {
		c.Status(http.StatusBadRequest)
		h.renderTemplate(c, "retrieve", getSecretData{
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
	h.renderTemplate(c, "retrieve", data)
}

// HandleRevealSecret handles secret revelation
func (h *HTTPHandlers) HandleRevealSecret(c *gin.Context) {
	var form RevealSecretForm
	if err := c.ShouldBind(&form); err != nil {
		h.HandleValidationError(c, err)
		return
	}

	finalization, err := base64.StdEncoding.DecodeString(form.Finalization)
	if err != nil {
		h.renderRevealResult(c, http.StatusBadRequest, fmt.Errorf("invalid finalization: %w", err), nil, nil)
		return
	}

	secretID, err := h.opaqueAuthUC.FinishLogin(form.SessionID, finalization)
	if err != nil {
		switch {
		case errors.Is(err, opaqueauth.ErrSessionNotFound):
			h.renderRevealResult(c, http.StatusBadRequest, errInvalidPinOrMissing, nil, nil)
		case errors.Is(err, opaqueauth.ErrSessionExpired):
			h.renderRevealResult(c, http.StatusBadRequest, errors.New("session expired, try again"), nil, nil)
		default:
			h.logger.Error().Err(err).Msg("opaque login finish")
			h.renderRevealResult(c, http.StatusBadRequest, errInvalidPinOrMissing, nil, nil)
		}
		return
	}

	if form.SecretID != "" && form.SecretID != secretID {
		h.renderRevealResult(c, http.StatusBadRequest, errInvalidPinOrMissing, nil, nil)
		return
	}

	secret, err := h.loadSecretUC.Execute(c.Request.Context(), secretID)
	if err != nil {
		if errors.Is(err, domain.ErrSecretNotFound) {
			h.renderRevealResult(c, http.StatusOK, errInvalidPinOrMissing, nil, nil)
			return
		}
		if errors.Is(err, domain.ErrSecretExpired) {
			h.renderRevealResult(c, http.StatusGone, errors.New("secret expired"), nil, nil)
			return
		}
		h.logger.Error().Err(err).Str("secret_id", secretID).Msg("load secret")
		h.renderRevealResult(c, http.StatusInternalServerError, errors.New("failed to load secret"), nil, nil)
		return
	}

	if err := h.deleteSecretUC.Execute(c.Request.Context(), secretID); err != nil {
		h.logger.Error().Err(err).Str("secret_id", secretID).Msg("delete secret after reveal")
	}

	h.renderRevealResult(c, http.StatusOK, nil, secret, secret.Payload)
}

func (h *HTTPHandlers) readUploadedFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	limit := int64(h.cfg.MaxPayloadBytes())
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

func normalizePayloadType(raw string, usedPlainText bool) domain.PayloadType {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(domain.PayloadTypeText):
		return domain.PayloadTypeText
	case string(domain.PayloadTypeFile):
		return domain.PayloadTypeFile
	}
	if usedPlainText {
		return domain.PayloadTypeText
	}
	return domain.PayloadTypeFile
}
