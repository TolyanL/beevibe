package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) UpdateVisibility(ctx context.Context, input dto.UpdateGameVisibilityInput) (dto.UpdateGameVisibilityOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game UpdateVisibility")
	defer span.End()

	var output dto.UpdateGameVisibilityOutput

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

	if domain.Visibility(input.Visibility) == domain.VisibilityPublic {
		err = game.Publish()
	} else {
		err = game.Unpublish()
	}

	if err != nil {
		return output, fmt.Errorf("game.Visibility: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.UpdateGameVisibility(ctx, game); err != nil {
			return fmt.Errorf("store.UpdateGameVisibility: %w", err)
		}

		return nil
	}); err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.Visibility = string(game.Visibility)

	return output, nil
}
