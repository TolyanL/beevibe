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

	CreateSession(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error
	GetSessionUserID(ctx context.Context, token string) (uuid.UUID, error)
	DeleteSession(ctx context.Context, token string) error

	InsertTokenOp(ctx context.Context, op domain.TokenOp) error
	ListTokenOps(ctx context.Context, userID uuid.UUID, limit int) ([]domain.TokenOp, error)

	ListUserGamesWithCost(ctx context.Context, userID uuid.UUID, limit, offset int) ([]dto.ProfileGame, error)
}

type Auth interface {
	HashPassword(password string) (string, error)
	CheckPassword(hash, password string) error
	GenerateToken() (string, error)
}

type Tokens struct {
	GrantSignup   int64
	GrantTopup    int64
	PriceGenerate int64
	PricePatch    int64
}

type UseCase struct {
	store  Store
	auth   Auth
	tokens Tokens
}

func New(store Store, auth Auth, tokens Tokens) *UseCase {
	return &UseCase{store: store, auth: auth, tokens: tokens}
}
