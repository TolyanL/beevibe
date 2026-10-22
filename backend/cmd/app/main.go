package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/TolyanL/beevibe/config"
	"github.com/TolyanL/beevibe/internal/app"
	"github.com/TolyanL/beevibe/pkg/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Debug().Err(err).Msg(".env not loaded")
	}

	c, err := config.New()
	if err != nil {
		log.Fatal().Err(err).Msg("config.New")
	}

	logger.Init(c.Logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, c); err != nil {
		log.Error().Err(err).Msg("app.Run")
	}
}
