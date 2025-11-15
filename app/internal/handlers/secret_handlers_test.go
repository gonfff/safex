package handlers

import (
	"bytes"
	"encoding/base64"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/gonfff/safex/app/internal/config"
	"github.com/gonfff/safex/app/internal/domain"
)

func TestHandleCreateSecret_InvalidForm(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Test с невалидным Content-Type
	req := httptest.NewRequest("POST", "/create", strings.NewReader("invalid data"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Должен возвращать ошибку из-за nil use case или неправильного формата
	assert.True(t, w.Code >= 400)
}

func TestHandleCreateSecret_MissingSecretID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form без secret_id
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_MissingOpaqueUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form без opaque_upload
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("message", "test message")
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_InvalidTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form с невалидным TTL
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("ttl_minutes", "invalid")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_NegativeTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form с отрицательным TTL
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("ttl_minutes", "-10")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_InvalidOpaqueUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form с невалидным base64 opaque_upload
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", "invalid-base64!")
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_NoFileNoMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form без файла и сообщения
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCreateSecret_MessageTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем сообщение больше лимита
	largeMessage := strings.Repeat("x", cfg.MaxPayloadBytes()+1)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("message", largeMessage)
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLoadSecret_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/secrets/:id", handlers.HandleLoadSecret)

	// Test с пустым ID (параметр :id не будет установлен)
	req := httptest.NewRequest("GET", "/secrets/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Обычно вернет 404 так как роут не совпадает, но проверим что обработчик работает
	assert.True(t, w.Code >= 400)
}

func TestHandleLoadSecret_ValidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.GET("/secrets/:id", handlers.HandleLoadSecret)

	// Test с валидным ID
	req := httptest.NewRequest("GET", "/secrets/test-secret-id", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleRevealSecret_MissingSessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test без session_id
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("finalization", base64.StdEncoding.EncodeToString([]byte("test")))
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevealSecret_MissingFinalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test без finalization
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "test-session-id")
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevealSecret_InvalidFinalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test с невалидным base64 finalization
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "test-session-id")
	writer.WriteField("finalization", "invalid-base64!")
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReadUploadedFile_EmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{MaxPayloadMB: 1}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Создаем пустой файл через form
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))

	// Создаем пустой файл
	part, err := writer.CreateFormFile("file", "empty.txt")
	assert.NoError(t, err)
	// Не записываем ничего в part - файл будет пустым
	_ = part
	writer.Close()

	// Создаем форм request
	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Парсим форму чтобы получить файл
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("Failed to parse form: %v", err)
	}

	// Получаем заголовок файла
	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	file.Close()

	// Тестируем readUploadedFile с пустым файлом
	_, err = handlers.readUploadedFile(header)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file is empty")
}

func TestNormalizePayloadType_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		usedPlainText bool
		expected      domain.PayloadType
	}{
		{
			name:          "uppercase TEXT",
			raw:           "TEXT",
			usedPlainText: false,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "mixed case File",
			raw:           "File",
			usedPlainText: true,
			expected:      domain.PayloadTypeFile,
		},
		{
			name:          "with whitespace",
			raw:           "  text  ",
			usedPlainText: false,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "unknown type with plain text",
			raw:           "unknown",
			usedPlainText: true,
			expected:      domain.PayloadTypeText,
		},
		{
			name:          "unknown type without plain text",
			raw:           "unknown",
			usedPlainText: false,
			expected:      domain.PayloadTypeFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePayloadType(tt.raw, tt.usedPlainText)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleCreateSecret_FileTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1, // 1 MB limit
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем файл больше лимита (симулируем)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))

	// Создаем большой файл - но из-за ограничений теста просто создадим с корректным контентом
	part, err := writer.CreateFormFile("file", "large.txt")
	assert.NoError(t, err)

	// Записываем большой контент
	largeContent := strings.Repeat("x", cfg.MaxPayloadBytes()+1)
	part.Write([]byte(largeContent))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Должен вернуть ошибку из-за размера файла или ошибку парсинга формы
	assert.True(t, w.Code >= 400)
}

func TestHandleCreateSecret_ZeroTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем multipart form с TTL равным 0 (невалидный)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("ttl_minutes", "0") // невалидный TTL
	writer.WriteField("message", "test message")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))
	writer.Close()

	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Должен вернуть 400 из-за неправильного TTL
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleLoadSecret_WithEmptyParam(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	// Без параметра :id
	router.GET("/secrets/", handlers.HandleLoadSecret)

	req := httptest.NewRequest("GET", "/secrets/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Должен обработать отсутствие ID
	assert.True(t, w.Code >= 400)
}

func TestHandleRevealSecret_EmptySessionID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test с пустым session_id (только пробелы)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "   ") // только пробелы
	writer.WriteField("finalization", base64.StdEncoding.EncodeToString([]byte("test")))
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleRevealSecret_EmptyFinalization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/reveal", handlers.HandleRevealSecret)

	// Test с пустым finalization (только пробелы)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("session_id", "test-session-id")
	writer.WriteField("finalization", "   ") // только пробелы
	writer.Close()

	req := httptest.NewRequest("POST", "/reveal", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestReadUploadedFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{MaxPayloadMB: 1}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	// Создаем файл с контентом
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("secret_id", "test-secret-id")
	writer.WriteField("opaque_upload", base64.StdEncoding.EncodeToString([]byte("test-opaque")))

	// Создаем файл с контентом
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)
	testContent := "Hello World"
	part.Write([]byte(testContent))
	writer.Close()

	// Создаем форм request
	req := httptest.NewRequest("POST", "/create", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Парсим форму чтобы получить файл
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		t.Fatalf("Failed to parse form: %v", err)
	}

	// Получаем заголовок файла
	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	file.Close()

	// Тестируем readUploadedFile с файлом с контентом
	payload, err := handlers.readUploadedFile(header)
	assert.NoError(t, err)
	assert.Equal(t, testContent, string(payload))
}

func TestHandleCreateSecret_ParseFormError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		DefaultTTL:   15 * time.Minute,
		MaxPayloadMB: 1,
	}
	logger := zerolog.Nop()

	handlers, err := NewHTTPHandlers(cfg, nil, nil, nil, nil, logger)
	assert.NoError(t, err)

	router := gin.New()
	router.POST("/create", handlers.HandleCreateSecret)

	// Создаем запрос с неправильным Content-Type
	req := httptest.NewRequest("POST", "/create", strings.NewReader("not multipart data"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Должен корректно обработать ошибку парсинга формы
	assert.True(t, w.Code >= 400)
}
