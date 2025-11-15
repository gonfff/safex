package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/gonfff/safex/app/internal/domain"
)

// Data types for templates
type homePageData struct {
	DefaultTTLMinutes  int
	MaxPayloadMB       int
	MaxPayloadBytes    int
	RateLimitPerMinute int
	Meta               pageMeta
}

type faqPageData struct {
	Meta pageMeta
}

type createResultData struct {
	Error     error
	Record    domain.Secret
	TTL       time.Duration
	MaxBytes  int
	SharePath string
	ShareURL  string
}

type revealResultData struct {
	Error         error
	Record        *domain.Secret
	PayloadBase64 string
	PayloadText   string
	IsText        bool
	InvalidPin    bool
}

type getSecretData struct {
	Title    string
	SecretID string
	Error    error
	Meta     pageMeta
}

type pageMeta struct {
	Canonical     string
	OGTitle       string
	OGDescription string
	OGType        string
	OGImage       string
}

// renderTemplate renders a template
func (h *HTTPHandlers) renderTemplate(c *gin.Context, name string, data any) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.templates.ExecuteTemplate(c.Writer, name, data); err != nil {
		h.logger.Error().Err(err).Str("template", name).Msg("render template")
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// renderCreateResult renders the secret creation result
func (h *HTTPHandlers) renderCreateResult(c *gin.Context, status int, renderErr error) {
	c.Status(status)
	h.renderTemplate(c, "createResult", createResultData{Error: renderErr})
}

// renderRevealResult renders the secret reveal result
func (h *HTTPHandlers) renderRevealResult(c *gin.Context, status int, renderErr error, secret *domain.Secret, payload []byte) {
	c.Status(status)
	data := revealResultData{
		Error: renderErr,
	}
	if renderErr != nil && renderErr.Error() == errInvalidPinOrMissing.Error() {
		data.InvalidPin = true
	}
	if secret != nil {
		data.Record = secret
		if secret.PayloadType == domain.PayloadTypeText {
			data.IsText = true
		}
	}
	if len(payload) > 0 {
		data.PayloadBase64 = base64.StdEncoding.EncodeToString(payload)
		if secret != nil && strings.HasPrefix(secret.ContentType, "text/") && utf8.Valid(payload) {
			data.PayloadText = string(payload)
			data.IsText = true
		}
	}
	h.renderTemplate(c, "revealResult", data)
}

// buildPageMeta builds page metadata
func (h *HTTPHandlers) buildPageMeta(r *http.Request, path, title, description string) pageMeta {
	if path == "" && r.URL != nil {
		path = r.URL.Path
	}
	if path == "" {
		path = "/"
	}
	return pageMeta{
		Canonical:     h.makeAbsoluteURL(r, path),
		OGTitle:       title,
		OGDescription: description,
		OGType:        "website",
	}
}

// makeShareURL creates a URL for sharing the secret
func (h *HTTPHandlers) makeShareURL(r *http.Request, id string) string {
	return h.makeAbsoluteURL(r, fmt.Sprintf("/secrets/%s", id))
}

// makeAbsoluteURL creates an absolute URL
func (h *HTTPHandlers) makeAbsoluteURL(r *http.Request, path string) string {
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
		addr := h.cfg.HTTPAddr
		if strings.HasPrefix(addr, ":") {
			host = "localhost" + addr
		} else {
			host = addr
		}
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, path)
}
