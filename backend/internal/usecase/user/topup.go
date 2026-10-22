package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) TopupTokens(ctx context.Context, input dto.TopupTokensInput) (dto.TopupTokensOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase user TopupTokens")
	defer span.End()

	var output dto.TopupTokensOutput

	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		user, err := u.store.GetUserByIDForUpdate(ctx, input.UserID)
		if err != nil {
			return fmt.Errorf("store.GetUserByIDForUpdate: %w", err)
		}

		if err := user.Credit(u.tokens.GrantTopup); err != nil {
			return fmt.Errorf("user.Credit: %w", err)
		}

		if err := u.store.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("store.UpdateUser: %w", err)
		}

		op, err := domain.NewTokenOp(user.ID, nil, u.tokens.GrantTopup, domain.ReasonTopup)
		if err != nil {
			return fmt.Errorf("domain.NewTokenOp: %w", err)
		}

		if err := u.store.InsertTokenOp(ctx, op); err != nil {
			return fmt.Errorf("store.InsertTokenOp: %w", err)
		}

		output.Balance = user.TokenBalance

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	return output, nil
}
