package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) ListUserGames(ctx context.Context, input dto.ListUserGamesInput) (dto.ListUserGamesOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game ListUserGames")
	defer span.End()

	var output dto.ListUserGamesOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	games, total, err := u.store.ListUserGames(ctx, input.UserID, dto.ListUserGamesInput{
		Status: input.Status,
		Sort:   input.Sort,
		Limit:  limit,
		Offset: input.Offset,
	})
	if err != nil {
		return output, fmt.Errorf("store.ListUserGames: %w", err)
	}

	output.Items = games
	output.Total = total

	return output, nil
}
