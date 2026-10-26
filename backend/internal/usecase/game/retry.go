package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) Retry(ctx context.Context, input dto.RetryGameInput) (dto.RetryGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Retry")
	defer span.End()

	var output dto.RetryGameOutput

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	if err := game.Retry(); err != nil {
		return output, fmt.Errorf("game.Retry: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.UpdateGameStatus(ctx, game.ID, game.Status, game.FailReason); err != nil {
			return fmt.Errorf("store.UpdateGameStatus: %w", err)
		}

		return nil
	}); err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.Status = string(game.Status)

	return output, nil
}
