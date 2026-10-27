package game

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/llm"
)

type Store interface {
	CreateGame(ctx context.Context, game domain.Game) error
	GetGame(ctx context.Context, gameID uuid.UUID) (domain.Game, error)
	GetGameBySlug(ctx context.Context, slug string) (domain.Game, error)
	DeleteGame(ctx context.Context, game domain.Game) error
	UpdateGameVisibility(ctx context.Context, game domain.Game) error
	UpdateGameMeta(ctx context.Context, game domain.Game) error
	UpdateGame(ctx context.Context, game domain.Game) error
	UpdateGameStatus(ctx context.Context, gameID uuid.UUID, status domain.Status, failReason string) error
	ListUserGames(ctx context.Context, userID uuid.UUID, filter dto.ListUserGamesInput) ([]dto.ProfileGame, int, error)
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

	SaveGameArtifact(ctx context.Context, gameID uuid.UUID, kind, content string) error
	ListGameArtifacts(ctx context.Context, gameID uuid.UUID) (map[string]string, error)

	InsertGenEvent(ctx context.Context, event domain.GenEvent) error
	ListGenEvents(ctx context.Context, gameID uuid.UUID, afterSeq int64) ([]domain.GenEvent, error)

	GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetUserByIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	InsertTokenOp(ctx context.Context, op domain.TokenOp) error

	InsertGameJob(ctx context.Context, job domain.GameJob) error
	ClaimGameJob(ctx context.Context) (domain.GameJob, error)

	InsertIdempotencyKey(ctx context.Context, key, operation string, userID uuid.UUID) (bool, error)
	GetIdempotencyResult(ctx context.Context, key, operation string) (string, bool, error)
	CompleteIdempotencyKey(ctx context.Context, key, operation, resultID string) error
	DeleteIdempotencyKey(ctx context.Context, key, operation string) error
}

type Validator interface {
	Check(ctx context.Context, html string) ([]domain.Issue, error)
	ExtractConfig(ctx context.Context, html string) (domain.GameConfig, error)
	ApplyConfig(ctx context.Context, html string, cfg domain.GameConfig) (string, error)
}

type Outbox interface {
	InsertOutboxEvent(ctx context.Context, event domain.Event) error
}

type Tokens struct {
	PriceGenerate int64
	PricePatch    int64
	PriceMeta     int64
}

type Reasoning struct {
	Analyze  llm.Effort
	Plan     llm.Effort
	Code     llm.Effort
	Assets   llm.Effort
	Review   llm.Effort
	Assemble llm.Effort
	Patch    llm.Effort
	Repair   llm.Effort
	Meta     llm.Effort
}

type Worker struct {
	Timeout         time.Duration
	MaxCallsPerGame int
	MaxFixAttempts  int
}

type runningTask struct {
	cancel  context.CancelFunc
	replied atomic.Bool
}

type taskKey struct{}

type UseCase struct {
	mu      sync.Mutex
	running map[uuid.UUID]*runningTask

	store     Store
	outbox    Outbox
	llm       llm.LLM
	validator Validator
	tokens    Tokens
	reasoning Reasoning
	worker    Worker
}

func New(
	store Store,
	outbox Outbox,
	llmClient llm.LLM,
	validator Validator,
	tokens Tokens,
	reasoning Reasoning,
	worker Worker,
) *UseCase {
	return &UseCase{
		running:   make(map[uuid.UUID]*runningTask),
		store:     store,
		outbox:    outbox,
		llm:       llmClient,
		validator: validator,
		tokens:    tokens,
		reasoning: reasoning,
		worker:    worker,
	}
}

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
