package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) ListGenEvents(ctx context.Context, input dto.ListGenEventsInput) (dto.ListGenEventsOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase ListGenEvents")
	defer span.End()

	var output dto.ListGenEventsOutput

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	events, err := u.store.ListGenEvents(ctx, input.GameID, input.AfterSeq)
	if err != nil {
		return output, fmt.Errorf("store.ListGenEvents: %w", err)
	}

	output.Events = events
	output.Status = string(game.Status)

	return output, nil
}
