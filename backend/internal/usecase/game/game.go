package game

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/llm"
)

// Store — порт доступа к данным игр, версий, оценок и игровых сессий.
type Store interface {
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

	UpsertRating(ctx context.Context, rating domain.Rating) error
	GetRating(ctx context.Context, gameID, userID uuid.UUID) (domain.Rating, error)
	GetRatingAggregate(ctx context.Context, gameID uuid.UUID) (dto.RatingAggregate, error)

	InsertPlaySession(ctx context.Context, session domain.PlaySession) error
	GetPlaytimeAggregate(ctx context.Context, gameID uuid.UUID) (dto.PlaytimeAggregate, error)

	InsertGenEvent(ctx context.Context, event domain.GenEvent) error
	ListGenEvents(ctx context.Context, gameID uuid.UUID, afterSeq int64) ([]domain.GenEvent, error)

	GetUserByIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	InsertTokenOp(ctx context.Context, op domain.TokenOp) error

	InsertIdempotencyKey(ctx context.Context, key, operation string, userID uuid.UUID) (bool, error)
	GetIdempotencyResult(ctx context.Context, key, operation string) (string, bool, error)
	CompleteIdempotencyKey(ctx context.Context, key, operation, resultID string) error
	DeleteIdempotencyKey(ctx context.Context, key, operation string) error
}

// Validator — порт проверки готовой игры. Реализация: internal/adapter/validator.
type Validator interface {
	Check(ctx context.Context, html string) ([]domain.Issue, error)
	ExtractConfig(ctx context.Context, html string) (domain.GameConfig, error)
	ApplyConfig(ctx context.Context, html string, cfg domain.GameConfig) (string, error)
}

// Outbox — порт записи события в outbox.
type Outbox interface {
	InsertOutboxEvent(ctx context.Context, event domain.Event) error
}

// Tokens — цены генерации и патча в токенах.
type Tokens struct {
	PriceGenerate int64
	PricePatch    int64
}

// Limits — лимиты токенов по агентам конвейера.
type Limits struct {
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

type UseCase struct {
	store     Store
	outbox    Outbox // TODO: писать lifecycle-события (domain.Event) в outbox — порт подключён, вызовов нет.
	llm       llm.LLM
	validator Validator
	tokens    Tokens
	limits    Limits
	worker    Worker
}

func New(store Store, outbox Outbox, llmClient llm.LLM, validator Validator, tokens Tokens, limits Limits, worker Worker) *UseCase {
	return &UseCase{
		store:     store,
		outbox:    outbox,
		llm:       llmClient,
		validator: validator,
		tokens:    tokens,
		limits:    limits,
		worker:    worker,
	}
}

// nextVersionNumber — следующий номер версии игры.
func (u *UseCase) nextVersionNumber(ctx context.Context, gameID uuid.UUID) (int, error) {
	latest, err := u.store.GetLatestVersion(ctx, gameID)
	if err != nil {
		if errors.Is(err, domain.ErrVersionNotFound) {
			return 1, nil
		}

		return 0, fmt.Errorf("store.GetLatestVersion: %w", err)
	}

	return latest.Number + 1, nil
}
