package user

import (
	"context"
	"fmt"
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

// SessionTTL — время жизни сессии.
const SessionTTL = 30 * 24 * time.Hour

func (u *UseCase) RegisterUser(ctx context.Context, input dto.RegisterUserInput) (dto.RegisterUserOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase user RegisterUser")
	defer span.End()

	var output dto.RegisterUserOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	hash, err := u.auth.HashPassword(input.Password)
	if err != nil {
		return output, fmt.Errorf("auth.HashPassword: %w", err)
	}

	user, err := domain.NewUser(input.Email, hash, u.tokens.GrantSignup)
	if err != nil {
		return output, fmt.Errorf("domain.NewUser: %w", err)
	}

	token, err := u.auth.GenerateToken()
	if err != nil {
		return output, fmt.Errorf("auth.GenerateToken: %w", err)
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("store.CreateUser: %w", err)
		}

		op, err := domain.NewTokenOp(user.ID, nil, u.tokens.GrantSignup, domain.ReasonSignupGrant)
		if err != nil {
			return fmt.Errorf("domain.NewTokenOp: %w", err)
		}

		if err := u.store.InsertTokenOp(ctx, op); err != nil {
			return fmt.Errorf("store.InsertTokenOp: %w", err)
		}

		if err := u.store.CreateSession(ctx, token, user.ID, time.Now().Add(SessionTTL)); err != nil {
			return fmt.Errorf("store.CreateSession: %w", err)
		}

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.User = dto.User{ID: user.ID, Email: string(user.Email), Balance: user.TokenBalance}
	output.SessionToken = token

	return output, nil
}
