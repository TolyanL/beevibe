package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) LoginUser(ctx context.Context, input dto.LoginUserInput) (dto.LoginUserOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase LoginUser")
	defer span.End()

	var output dto.LoginUserOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	user, err := u.store.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return output, domain.ErrUnauthorized
		}

		return output, fmt.Errorf("store.GetUserByEmail: %w", err)
	}

	if err := u.auth.CheckPassword(string(user.PasswordHash), input.Password); err != nil {
		return output, domain.ErrUnauthorized
	}

	token, err := u.auth.GenerateToken()
	if err != nil {
		return output, fmt.Errorf("auth.GenerateToken: %w", err)
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
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
