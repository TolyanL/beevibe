package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) UpdateProfile(ctx context.Context, input dto.UpdateProfileInput) (dto.UpdateProfileOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase user UpdateProfile")
	defer span.End()

	var output dto.UpdateProfileOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	user, err := u.store.GetUserByID(ctx, input.UserID)
	if err != nil {
		return output, fmt.Errorf("store.GetUserByID: %w", err)
	}

	if err := user.SetDisplayName(input.DisplayName); err != nil {
		return output, fmt.Errorf("user.SetDisplayName: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.UpdateDisplayName(ctx, user); err != nil {
			return fmt.Errorf("store.UpdateDisplayName: %w", err)
		}

		return nil
	}); err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.User = dto.User{
		ID:          user.ID,
		Email:       string(user.Email),
		DisplayName: user.DisplayName,
		Balance:     user.TokenBalance,
	}

	return output, nil
}
