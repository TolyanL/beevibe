package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) GetMe(ctx context.Context, input dto.GetMeInput) (dto.GetMeOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase GetMe")
	defer span.End()

	var output dto.GetMeOutput

	user, err := u.store.GetUserByID(ctx, input.UserID)
	if err != nil {
		return output, fmt.Errorf("store.GetUserByID: %w", err)
	}

	output.User = dto.User{ID: user.ID, Email: string(user.Email), Balance: user.TokenBalance}

	return output, nil
}
