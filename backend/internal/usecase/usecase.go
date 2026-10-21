package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

// Store — порт доступа к данным. Реализация: internal/adapter/postgres.
type Store interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetUserByIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error

	CreateSession(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error
	GetSessionUserID(ctx context.Context, token string) (uuid.UUID, error)
	DeleteSession(ctx context.Context, token string) error

	CreateGame(ctx context.Context, game domain.Game) error
	GetGame(ctx context.Context, gameID uuid.UUID) (domain.Game, error)
	UpdateGame(ctx context.Context, game domain.Game) error
	UpdateGameStatus(ctx context.Context, gameID uuid.UUID, status domain.Status, failReason string) error
	ListUserGamesWithCost(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.ProfileGame, error)
	ClaimQueuedGame(ctx context.Context) (domain.Game, error)

	CreateVersion(ctx context.Context, version domain.Version) error
	GetVersion(ctx context.Context, gameID uuid.UUID, number int) (domain.Version, error)
	GetLatestVersion(ctx context.Context, gameID uuid.UUID) (domain.Version, error)
	ListVersions(ctx context.Context, gameID uuid.UUID) ([]domain.Version, error)

	InsertTokenOp(ctx context.Context, op domain.TokenOp) error
	ListTokenOps(ctx context.Context, userID uuid.UUID, limit int) ([]domain.TokenOp, error)

	UpsertRating(ctx context.Context, rating domain.Rating) error
	GetRating(ctx context.Context, gameID, userID uuid.UUID) (domain.Rating, error)
	GetRatingAggregate(ctx context.Context, gameID uuid.UUID) (dto.RatingAggregate, error)

	InsertPlaySession(ctx context.Context, session domain.PlaySession) error
	GetPlaytimeAggregate(ctx context.Context, gameID uuid.UUID) (dto.PlaytimeAggregate, error)

	InsertGenEvent(ctx context.Context, event domain.GenEvent) error
	ListGenEvents(ctx context.Context, gameID uuid.UUID, afterSeq int64) ([]domain.GenEvent, error)

	ListCatalog(ctx context.Context, filter dto.CatalogFilter) ([]dto.CatalogCard, error)

	InsertIdempotencyKey(ctx context.Context, key, operation string, userID uuid.UUID) (bool, error)
	GetIdempotencyResult(ctx context.Context, key, operation string) (string, bool, error)
	CompleteIdempotencyKey(ctx context.Context, key, operation, resultID string) error
	DeleteIdempotencyKey(ctx context.Context, key, operation string) error
}

// LLM — порт чат-комплишена. Реализация: internal/adapter/llm.
type LLM interface {
	Complete(ctx context.Context, req dto.LLMRequest) (dto.LLMResponse, error)
	Prompt(name string) (string, error)
	PromptVersion() string
}

// Validator — порт проверки готовой игры. Реализация: internal/adapter/validator.
type Validator interface {
	Check(ctx context.Context, html string) ([]domain.Issue, error)
	ExtractConfig(ctx context.Context, html string) (domain.GameConfig, error)
	ApplyConfig(ctx context.Context, html string, cfg domain.GameConfig) (string, error)
}

// Auth — порт кредов и сессий. Реализация: internal/adapter/auth.
type Auth interface {
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
	GenerateToken() (string, error)
}

// Tokens — цены и гранты в токенах.
type Tokens struct {
	GrantSignup   int64
	GrantTopup    int64
	PriceGenerate int64
	PricePatch    int64
}

// LLMLimits — лимиты токенов по агентам конвейера.
type LLMLimits struct {
	MaxTokensAnalyze  int
	MaxTokensPlan     int
	MaxTokensCode     int
	MaxTokensAssets   int
	MaxTokensReview   int
	MaxTokensAssemble int
	MaxTokensPatch    int
	MaxTokensRepair   int
}

// Worker — настройки конвейера генерации.
type Worker struct {
	Timeout         time.Duration
	MaxCallsPerGame int
	MaxFixAttempts  int
}

// App — метаданные сервиса.
type App struct {
	Name    string
	Version string
}

// Config — конфигурация слоя юзкейсов.
type Config struct {
	Tokens Tokens
	Limits LLMLimits
	Worker Worker
	App    App
}

type UseCase struct {
	store     Store
	llm       LLM
	validator Validator
	auth      Auth
	tokens    Tokens
	limits    LLMLimits
	worker    Worker
	app       App
}

func New(store Store, llm LLM, validator Validator, auth Auth, cfg Config) *UseCase {
	return &UseCase{
		store:     store,
		llm:       llm,
		validator: validator,
		auth:      auth,
		tokens:    cfg.Tokens,
		limits:    cfg.Limits,
		worker:    cfg.Worker,
		app:       cfg.App,
	}
}
