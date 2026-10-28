package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) Patch(ctx context.Context, input dto.PatchGameInput) (dto.PatchGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Patch")
	defer span.End()

	var output dto.PatchGameOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	if input.IdempotencyKey != "" {
		_, replayed, err := u.claimIdempotencyKey(ctx, input.IdempotencyKey, operationPatchGame, input.UserID)
		if err != nil {
			return output, err
		}

		if replayed {
			output.Status = string(domain.StatusPatching)

			return output, nil
		}
	}

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	if game.Status != domain.StatusReady {
		return output, domain.ErrGenerationInProgress
	}

	if _, err := u.store.GetLatestVersion(ctx, input.GameID); err != nil {
		return output, fmt.Errorf("store.GetLatestVersion: %w", err)
	}

	job, err := domain.NewGameJob(game.ID, game.OwnerID, domain.JobKindPatch, input.Text)
	if err != nil {
		return output, fmt.Errorf("domain.NewGameJob: %w", err)
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		user, err := u.store.GetUserByIDForUpdate(ctx, input.UserID)
		if err != nil {
			return fmt.Errorf("store.GetUserByIDForUpdate: %w", err)
		}

		if err := user.Debit(u.tokens.PricePatch); err != nil {
			return fmt.Errorf("user.Debit: %w", err)
		}

		if err := u.store.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("store.UpdateUser: %w", err)
		}

		op, err := domain.NewTokenOp(user.ID, &input.GameID, -u.tokens.PricePatch, domain.ReasonPatch)
		if err != nil {
			return fmt.Errorf("domain.NewTokenOp: %w", err)
		}

		if err := u.store.InsertTokenOp(ctx, op); err != nil {
			return fmt.Errorf("store.InsertTokenOp: %w", err)
		}

		if err := u.store.InsertGameJob(ctx, job); err != nil {
			return fmt.Errorf("store.InsertGameJob: %w", err)
		}

		return u.markPatching(ctx, &game, job.Kind)
	})
	if err != nil {
		u.releaseIdempotencyKey(ctx, input.IdempotencyKey, operationPatchGame)

		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if err := u.completeIdempotencyKey(ctx, input.IdempotencyKey, operationPatchGame, input.GameID.String()); err != nil {
		return output, err
	}

	output.Status = string(domain.StatusPatching)

	return output, nil
}

func (u *UseCase) patchHistory(ctx context.Context, gameID uuid.UUID) (string, error) {
	versions, err := u.store.ListVersions(ctx, gameID)
	if err != nil {
		return "", fmt.Errorf("store.ListVersions: %w", err)
	}

	history := ""

	for _, version := range versions {
		if version.Source != domain.VersionSourcePatch {
			continue
		}

		history += "- " + version.Note + "\n"
	}

	return history, nil
}

func (u *UseCase) refundInTx(ctx context.Context, userID, gameID uuid.UUID, amount int64) error {
	user, err := u.store.GetUserByIDForUpdate(ctx, userID)
	if err != nil {
		return fmt.Errorf("store.GetUserByIDForUpdate: %w", err)
	}

	if err := user.Credit(amount); err != nil {
		return fmt.Errorf("user.Credit: %w", err)
	}

	if err := u.store.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("store.UpdateUser: %w", err)
	}

	op, err := domain.NewTokenOp(user.ID, &gameID, amount, domain.ReasonRefund)
	if err != nil {
		return fmt.Errorf("domain.NewTokenOp: %w", err)
	}

	if err := u.store.InsertTokenOp(ctx, op); err != nil {
		return fmt.Errorf("store.InsertTokenOp: %w", err)
	}

	return nil
}
