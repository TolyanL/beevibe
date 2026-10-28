package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) Delete(ctx context.Context, input dto.DeleteGameInput) (dto.DeleteGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Delete")
	defer span.End()

	var output dto.DeleteGameOutput

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	if err := game.Delete(); err != nil {
		return output, fmt.Errorf("game.Delete: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.DeleteGame(ctx, game); err != nil {
			return fmt.Errorf("store.DeleteGame: %w", err)
		}

		return nil
	}); err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.Deleted = true

	return output, nil
}
