package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gonfff/safex/app/internal/config"
	"github.com/gonfff/safex/app/internal/handlers"
	"github.com/gonfff/safex/app/internal/infrastructure/middleware"
	"github.com/gonfff/safex/app/web"
)

// Server HTTP server wrapper with clean architecture
type Server struct {
	cfg      config.Config
	handlers *handlers.HTTPHandlers
	engine   *gin.Engine
	httpSrv  *http.Server
	logger   zerolog.Logger
}

// NewServer creates a new HTTP server
func NewServer(cfg config.Config, handlers *handlers.HTTPHandlers, logger zerolog.Logger) (*Server, error) {
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
	engine.Use(gin.Recovery(), middleware.ZerologMiddleware(logger))
	engine.StaticFS("/static", http.FS(staticFS))

	var limiter *middleware.RateLimiter
	if cfg.RequestsPerMinute > 0 {
		limiter = middleware.NewRateLimiter(cfg.RequestsPerMinute, time.Minute)
		engine.Use(middleware.RateLimitMiddleware(limiter, logger))
	}

	s := &Server{
		cfg:      cfg,
		handlers: handlers,
		engine:   engine,
		logger:   logger,
	}
	s.registerRoutes()
	return s, nil
}

// Start starts the HTTP server
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
	s.engine.GET("/healthz", s.handlers.HandleHealth)
	s.engine.GET("/", s.handlers.HandleHome)
	s.engine.GET("/faq", s.handlers.HandleFAQ)
	s.engine.POST("/opaque/register/start", s.handlers.HandleOpaqueRegisterStart)
	s.engine.POST("/opaque/login/start", s.handlers.HandleOpaqueLoginStart)
	s.engine.POST("/secrets", s.handlers.HandleCreateSecret)
	s.engine.GET("/secrets/:id", s.handlers.HandleLoadSecret)
	s.engine.POST("/secrets/reveal", s.handlers.HandleRevealSecret)
}
