package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gonfff/safex/app/internal/adapters"
	"github.com/gonfff/safex/app/internal/config"
	"github.com/gonfff/safex/app/internal/handlers"
	"github.com/gonfff/safex/app/internal/infrastructure"
	"github.com/gonfff/safex/app/internal/opaqueauth"
	"github.com/gonfff/safex/app/internal/storage"
	"github.com/gonfff/safex/app/internal/usecases"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = logger

	cfg := config.Load()
	if err := cfg.MustValidate(); err != nil {
		logger.Fatal().Err(err).Msg("invalid config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize storage layer
	blobStore, err := storage.NewBlobStore(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init blob store")
	}
	metaStore, err := storage.NewMetadataStore(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init metadata store")
	}
	defer closeIfPossible(metaStore)
	defer closeIfPossible(blobStore)

	// Initialize OPAQUE manager
	opaqueMgr, err := opaqueauth.NewManager(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init opaque manager")
	}

	// Create adapters (Infrastructure layer -> Domain layer)
	secretRepo := adapters.NewSecretRepositoryAdapter(metaStore)
	blobRepo := adapters.NewBlobRepositoryAdapter(blobStore)
	opaqueAuthService := adapters.NewOpaqueAuthServiceAdapter(opaqueMgr)

	// Create use cases (Application layer)
	createSecretUC := usecases.NewCreateSecretUseCase(secretRepo, blobRepo, logger)
	loadSecretUC := usecases.NewLoadSecretUseCase(secretRepo, blobRepo, logger)
	deleteSecretUC := usecases.NewDeleteSecretUseCase(secretRepo, blobRepo, logger)
	opaqueAuthUC := usecases.NewOpaqueAuthUseCase(opaqueAuthService, logger)

	// Create handlers (Presentation layer)
	httpHandlers, err := handlers.NewHTTPHandlers(
		cfg,
		createSecretUC,
		loadSecretUC,
		deleteSecretUC,
		opaqueAuthUC,
		logger,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("init handlers")
	}

	// Create and start server (Infrastructure layer)
	srv, err := infrastructure.NewServer(cfg, httpHandlers, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("init server")
	}

	if err := srv.Start(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server stopped")
	}
}

func closeIfPossible(v any) {
	if closer, ok := v.(interface{ Close() error }); ok {
		_ = closer.Close()
	}
}
