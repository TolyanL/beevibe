package app

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"

	"github.com/TolyanL/beevibe/config"
	"github.com/TolyanL/beevibe/internal/adapter/auth"
	"github.com/TolyanL/beevibe/internal/adapter/postgres"
	"github.com/TolyanL/beevibe/internal/adapter/validator"
	"github.com/TolyanL/beevibe/internal/controller/http/public"
	v1 "github.com/TolyanL/beevibe/internal/controller/http/v1"
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

	store := postgres.New(pool)
	llmClient := llm.New(c.LLM)
	validatorAdapter := validator.New(c.Validator)
	authAdapter := auth.New(c.Auth)

	uc := usecase.New(store, llmClient, validatorAdapter, authAdapter, usecase.Config{
		Tokens: usecase.Tokens{
			GrantSignup:   c.Tokens.GrantSignup,
			GrantTopup:    c.Tokens.GrantTopup,
			PriceGenerate: c.Tokens.PriceGenerate,
			PricePatch:    c.Tokens.PricePatch,
		},
		Limits: usecase.LLMLimits{
			MaxTokensAnalyze:  c.LLMToken.Analyzer,
			MaxTokensPlan:     c.LLMToken.Planner,
			MaxTokensCode:     c.LLMToken.Coder,
			MaxTokensAssets:   c.LLMToken.Assets,
			MaxTokensReview:   c.LLMToken.Reviewer,
			MaxTokensAssemble: c.LLMToken.Assembler,
			MaxTokensPatch:    c.LLMToken.Patch,
			MaxTokensRepair:   c.LLMToken.Repair,
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

	server := httpserver.New(c.HTTP, root)

	worker := NewWorker(uc, c.Worker.Concurrency)
	worker.Start(ctx)

	errCh := make(chan error, 1)

	go func() {
		errCh <- server.Start()
	}()

	select {
	case err := <-errCh:
		worker.Close()

		return fmt.Errorf("server.Start: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), c.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		worker.Close()

		return fmt.Errorf("server.Shutdown: %w", err)
	}

	worker.Close()

	return nil
}
