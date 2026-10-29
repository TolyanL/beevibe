package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const (
	reviewAccept     = "accept"
	reviewRegenerate = "regenerate"
)

func (u *UseCase) ResolveReview(ctx context.Context, input dto.ResolveReviewInput) (dto.ResolveReviewOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game ResolveReview")
	defer span.End()

	var output dto.ResolveReviewOutput

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

	if game.Status != domain.StatusNeedsDecision {
		return output, fmt.Errorf("%w: %s is not waiting for a review decision", domain.ErrInvalidTransition, game.Status)
	}

	resume := domain.StatusAssembling
	if input.Action == reviewRegenerate {
		resume = domain.StatusCoding
	}

	if err := game.ScheduleResume(resume); err != nil {
		return output, fmt.Errorf("game.ScheduleResume: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.UpdateGameStatus(ctx, game.ID, game.Status, game.FailReason); err != nil {
			return fmt.Errorf("store.UpdateGameStatus: %w", err)
		}

		if err := u.store.SaveGameResumeStage(ctx, game.ID, resume); err != nil {
			return fmt.Errorf("store.SaveGameResumeStage: %w", err)
		}

		return nil
	}); err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.Status = string(game.Status)
	output.ResumeFrom = string(resume)

	return output, nil
}
