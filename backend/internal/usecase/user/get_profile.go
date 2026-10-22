package user

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

const profileGamesLimit = 50

func (u *UseCase) GetProfile(ctx context.Context, input dto.GetProfileInput) (dto.GetProfileOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase user GetProfile")
	defer span.End()

	var output dto.GetProfileOutput

	user, err := u.store.GetUserByID(ctx, input.UserID)
	if err != nil {
		return output, fmt.Errorf("store.GetUserByID: %w", err)
	}

	games, err := u.store.ListUserGamesWithCost(ctx, input.UserID, profileGamesLimit, 0)
	if err != nil {
		return output, fmt.Errorf("store.ListUserGamesWithCost: %w", err)
	}

	ops, err := u.store.ListTokenOps(ctx, input.UserID, profileGamesLimit)
	if err != nil {
		return output, fmt.Errorf("store.ListTokenOps: %w", err)
	}

	output.Balance = user.TokenBalance
	output.Games = games
	output.TokenOps = make([]dto.ProfileTokenOp, 0, len(ops))
	output.Prices = dto.ProfilePrices{
		Generate: u.tokens.PriceGenerate,
		Patch:    u.tokens.PricePatch,
	}

	for _, op := range ops {
		output.TokenOps = append(output.TokenOps, dto.ProfileTokenOp{
			GameID:    op.GameID,
			Delta:     op.Delta,
			Reason:    string(op.Reason),
			CreatedAt: op.CreatedAt,
		})
	}

	return output, nil
}
