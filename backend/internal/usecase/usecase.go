package usecase

import (
	"github.com/TolyanL/beevibe/internal/usecase/catalog"
	"github.com/TolyanL/beevibe/internal/usecase/event"
	"github.com/TolyanL/beevibe/internal/usecase/game"
	"github.com/TolyanL/beevibe/internal/usecase/outbox"
	"github.com/TolyanL/beevibe/internal/usecase/user"
	"github.com/TolyanL/beevibe/pkg/llm"
)

type Tokens = user.Tokens

type LLMReasoning = game.Reasoning

type Worker = game.Worker

type App struct {
	Name    string
	Version string
}

type Config struct {
	Tokens    Tokens
	Reasoning LLMReasoning
	Worker    Worker
	App       App
}

type Store interface {
	user.Store
	game.Store
	catalog.Store
	outbox.Store
}

type UseCase struct {
	User    *user.UseCase
	Game    *game.UseCase
	Catalog *catalog.UseCase
	Outbox  *outbox.UseCase
	Event   *event.UseCase

	app App
}

func New(store Store, llmClient llm.LLM, validator game.Validator, auth user.Auth, kafka outbox.Kafka, cfg Config) *UseCase {
	return &UseCase{
		User: user.New(store, auth, cfg.Tokens),
		Game: game.New(
			store,
			store,
			llmClient,
			validator,
			game.Tokens{
				PriceGenerate: cfg.Tokens.PriceGenerate,
				PricePatch:    cfg.Tokens.PricePatch,
				PriceMeta:     cfg.Tokens.PriceMeta,
			},
			cfg.Reasoning,
			cfg.Worker,
		),
		Catalog: catalog.New(store),
		Outbox:  outbox.New(store, kafka),
		Event:   event.New(store),
		app:     cfg.App,
	}
}
