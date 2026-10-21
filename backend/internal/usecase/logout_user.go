package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) LogoutUser(ctx context.Context, input dto.LogoutUserInput) (dto.LogoutUserOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase LogoutUser")
	defer span.End()

	var output dto.LogoutUserOutput

	if input.SessionToken == "" {
		return output, nil
	}

	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.DeleteSession(ctx, input.SessionToken); err != nil {
			return fmt.Errorf("store.DeleteSession: %w", err)
		}

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	return output, nil
}
