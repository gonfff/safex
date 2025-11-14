package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gonfff/safex/backend/internal/config"
	"github.com/gonfff/safex/backend/internal/opaqueauth"
	"github.com/gonfff/safex/backend/internal/secret"
	"github.com/gonfff/safex/backend/internal/server"
	"github.com/gonfff/safex/backend/internal/storage"
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

	svc := secret.NewService(blobStore, metaStore, logger)
	opaqueMgr, err := opaqueauth.NewManager(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("init opaque manager")
	}

	srv, err := server.New(cfg, svc, opaqueMgr, logger)
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
