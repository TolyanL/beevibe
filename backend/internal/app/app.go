package app

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/TolyanL/beevibe/internal/adapter/auth"
	"github.com/TolyanL/beevibe/internal/adapter/kafka"
	"github.com/TolyanL/beevibe/internal/adapter/postgres"
	"github.com/TolyanL/beevibe/internal/adapter/validator"
	"github.com/TolyanL/beevibe/internal/config"
	"github.com/TolyanL/beevibe/internal/controller/http/public"
	"github.com/TolyanL/beevibe/internal/controller/http/v1"
	"github.com/TolyanL/beevibe/internal/controller/kafka_consumer"
	"github.com/TolyanL/beevibe/internal/usecase"
	"github.com/TolyanL/beevibe/pkg/httpserver"
	"github.com/TolyanL/beevibe/pkg/llm"
	pgpool "github.com/TolyanL/beevibe/pkg/postgres"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func Run(ctx context.Context, c config.Config) error {
	pool, err := pgpool.New(ctx, c.Postgres)
	if err != nil {
		return fmt.Errorf("pgpool.New: %w", err)
	}

	defer pool.Close()

	transaction.Init(pool)

	log.Info().Str("model", c.LLM.Model).Str("llm_base_url", c.LLM.BaseURL).Msg("gamegen configured")

	store := postgres.New(pool)
	llmClient := llm.New(c.LLM)
	validatorAdapter := validator.New(c.Validator)
	authAdapter := auth.New(c.Auth)
	producer := kafka.New(c.Kafka)

	uc := usecase.New(store, llmClient, validatorAdapter, authAdapter, producer, usecase.Config{
		Tokens: usecase.Tokens{
			GrantSignup:   c.Tokens.GrantSignup,
			GrantTopup:    c.Tokens.GrantTopup,
			PriceGenerate: c.Tokens.PriceGenerate,
			PricePatch:    c.Tokens.PricePatch,
		},
		Reasoning: usecase.LLMReasoning{
			Analyze:  c.LLMReasoning.Analyzer,
			Plan:     c.LLMReasoning.Planner,
			Code:     c.LLMReasoning.Coder,
			Assets:   c.LLMReasoning.Assets,
			Review:   c.LLMReasoning.Reviewer,
			Assemble: c.LLMReasoning.Assembler,
			Patch:    c.LLMReasoning.Patch,
			Repair:   c.LLMReasoning.Repair,
		},
		Worker: usecase.Worker{
			Timeout:         c.Worker.Timeout,
			MaxCallsPerGame: c.Worker.MaxCallsPerGame,
			MaxFixAttempts:  c.Worker.MaxFixAttempts,
		},
		App: usecase.App{
			Name:    c.App.Name,
			Version: c.App.Version,
		},
	})

	root := chi.NewRouter()
	root.Mount("/api", v1.NewRouter(v1.New(uc)))
	root.Mount("/", public.NewRouter(public.New(uc)))

	serverCtx, cancelServer := context.WithCancel(context.Background())
	defer cancelServer()

	server := httpserver.New(c.HTTP, root, serverCtx)

	generationWorker := NewWorker(uc, c.Worker.Concurrency)
	generationWorker.Start(ctx)

	outboxWorker := NewOutboxWorker(uc, outboxPollInterval)
	outboxWorker.Start(ctx)

	consumer := kafka_consumer.New(c.Kafka, uc.Event)
	consumer.Start(ctx)

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Start()
	}()

	select {
	case err := <-errCh:
		closeAll(generationWorker, outboxWorker, consumer)

		return fmt.Errorf("server.Start: %w", err)
	case <-ctx.Done():
	}

	cancelServer()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), c.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		closeAll(generationWorker, outboxWorker, consumer)

		return fmt.Errorf("server.Shutdown: %w", err)
	}

	closeAll(generationWorker, outboxWorker, consumer)

	if err := producer.Close(); err != nil {
		return fmt.Errorf("producer.Close: %w", err)
	}

	return nil
}

func closeAll(generationWorker *Worker, outboxWorker *OutboxWorker, consumer *kafka_consumer.Consumer) {
	generationWorker.Close()
	outboxWorker.Close()

	if err := consumer.Close(); err != nil {
		log.Error().Err(err).Msg("kafka consumer close failed")
	}
}
