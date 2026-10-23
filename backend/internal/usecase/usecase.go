package usecase

import (
	"github.com/TolyanL/beevibe/internal/usecase/catalog"
	"github.com/TolyanL/beevibe/internal/usecase/event"
	"github.com/TolyanL/beevibe/internal/usecase/game"
	"github.com/TolyanL/beevibe/internal/usecase/outbox"
	"github.com/TolyanL/beevibe/internal/usecase/user"
	"github.com/TolyanL/beevibe/pkg/llm"
)

// Tokens — цены и гранты в токенах.
type Tokens = user.Tokens

// LLMReasoning — уровни рассуждения модели по агентам конвейера.
type LLMReasoning = game.Reasoning

// Worker — настройки конвейера генерации.
type Worker = game.Worker

// App — метаданные сервиса.
type App struct {
	Name    string
	Version string
}

// Config — конфигурация слоя юзкейсов.
type Config struct {
	Tokens    Tokens
	Reasoning LLMReasoning
	Worker    Worker
	App       App
}

// Store — порт доступа к данным. Реализация: internal/adapter/postgres.
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
