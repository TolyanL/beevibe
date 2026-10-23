package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) Rate(ctx context.Context, input dto.RateGameInput) (dto.RateGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Rate")
	defer span.End()

	var output dto.RateGameOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	rating, err := domain.NewRating(input.GameID, input.UserID, domain.RatingValue(input.Value))
	if err != nil {
		return output, fmt.Errorf("domain.NewRating: %w", err)
	}

	if err := u.store.UpsertRating(ctx, rating); err != nil {
		return output, fmt.Errorf("store.UpsertRating: %w", err)
	}

	aggregate, err := u.store.GetRatingAggregate(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetRatingAggregate: %w", err)
	}

	output.Rating = input.Value
	output.Aggregate = aggregate

	return output, nil
}
