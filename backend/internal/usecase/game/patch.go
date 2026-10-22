package game

import (
	"context"
	"errors"
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

	latest, err := u.store.GetLatestVersion(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetLatestVersion: %w", err)
	}

	history, err := u.patchHistory(ctx, input.GameID)
	if err != nil {
		return output, err
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

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	system, err := u.llm.Prompt("patch")
	if err != nil {
		return output, errors.Join(fmt.Errorf("llm.Prompt: %w", err), u.refund(ctx, input.UserID, input.GameID, u.tokens.PricePatch))
	}

	p := &pipeline{}

	raw, err := u.complete(ctx, p, system, buildPatchPrompt(latest.HTML, input.Text, history), temperatureCode, u.maxTokensPatch())
	if err != nil {
		return output, errors.Join(fmt.Errorf("u.complete: %w", err), u.refund(ctx, input.UserID, input.GameID, u.tokens.PricePatch))
	}

	issues, err := u.validator.Check(ctx, raw)
	if err != nil {
		return output, errors.Join(fmt.Errorf("validator.Check: %w", err), u.refund(ctx, input.UserID, input.GameID, u.tokens.PricePatch))
	}

	if len(issues) > 0 {
		return output, errors.Join(
			fmt.Errorf("%w: %s", domain.ErrValidationFailed, issues[0].Message),
			u.refund(ctx, input.UserID, input.GameID, u.tokens.PricePatch),
		)
	}

	number, err := u.nextVersionNumber(ctx, input.GameID)
	if err != nil {
		return output, err
	}

	version, err := domain.NewVersion(input.GameID, number, raw, domain.VersionSourcePatch)
	if err != nil {
		return output, fmt.Errorf("domain.NewVersion: %w", err)
	}

	version.Note = input.Text
	version.TokensCost = p.tokens
	version.ModelCalls = p.calls
	version.PromptVersion = u.llm.PromptVersion()

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.CreateVersion(ctx, version); err != nil {
			return fmt.Errorf("store.CreateVersion: %w", err)
		}

		if err := u.publishEvent(ctx, game, domain.EventGamePatched, &version.ID, p.tokens, ""); err != nil {
			return fmt.Errorf("u.publishEvent: %w", err)
		}

		return nil
	})
	if err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.VersionID = version.ID
	output.TokensSpent = p.tokens

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

func (u *UseCase) refund(ctx context.Context, userID, gameID uuid.UUID, amount int64) error {
	err := transaction.Wrap(ctx, func(ctx context.Context) error {
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
	})
	if err != nil {
		return fmt.Errorf("transaction.Wrap: %w", err)
	}

	return nil
}
