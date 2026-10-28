package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

type Store interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetUserByIDForUpdate(ctx context.Context, userID uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) error
	UpdateDisplayName(ctx context.Context, user domain.User) error

	CreateSession(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error
	GetSessionUserID(ctx context.Context, token string) (uuid.UUID, error)
	DeleteSession(ctx context.Context, token string) error

	InsertTokenOp(ctx context.Context, op domain.TokenOp) error
	ListTokenOps(ctx context.Context, userID uuid.UUID, filter dto.ListTokenOpsInput) ([]domain.TokenOp, int, error)
	SumTopupsSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error)

	InsertIdempotencyKey(ctx context.Context, key, operation string, userID uuid.UUID) (bool, error)
	GetIdempotencyResult(ctx context.Context, key, operation string) (string, bool, error)
	CompleteIdempotencyKey(ctx context.Context, key, operation, resultID string) error
	DeleteIdempotencyKey(ctx context.Context, key, operation string) error

	ListUserGames(ctx context.Context, userID uuid.UUID, filter dto.ListUserGamesInput) ([]dto.ProfileGame, int, error)
}

type Auth interface {
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
	GenerateToken() (string, error)
}

type Tokens struct {
	GrantSignup   int64
	PriceGenerate int64
	PricePatch    int64
	PriceMeta     int64
	TopupMin      int64
	TopupMax      int64
	TopupDailyCap int64
	DemoTopup     bool
}

type UseCase struct {
	store  Store
	auth   Auth
	tokens Tokens
}

func New(store Store, auth Auth, tokens Tokens) *UseCase {
	return &UseCase{store: store, auth: auth, tokens: tokens}
}
