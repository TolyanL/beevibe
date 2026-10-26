package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) Fix(ctx context.Context, input dto.FixGameInput) (dto.FixGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Fix")
	defer span.End()

	var output dto.FixGameOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	if input.IdempotencyKey != "" {
		_, replayed, err := u.claimIdempotencyKey(ctx, input.IdempotencyKey, operationFixGame, input.UserID)
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

	job, err := domain.NewGameJob(game.ID, game.OwnerID, domain.JobKindFix, input.Error)
	if err != nil {
		return output, fmt.Errorf("domain.NewGameJob: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.InsertGameJob(ctx, job); err != nil {
			return fmt.Errorf("store.InsertGameJob: %w", err)
		}

		return u.markPatching(ctx, &game, job.Kind)
	}); err != nil {
		u.releaseIdempotencyKey(ctx, input.IdempotencyKey, operationFixGame)

		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if err := u.completeIdempotencyKey(ctx, input.IdempotencyKey, operationFixGame, input.GameID.String()); err != nil {
		return output, err
	}

	output.Status = string(domain.StatusPatching)

	return output, nil
}

func (u *UseCase) repair(ctx context.Context, p *pipeline, html string, issues []domain.Issue) (string, error) {
	system, err := u.llm.Prompt("repair")
	if err != nil {
		return "", fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.complete(ctx, p, system, buildRepairPrompt(html, issues), temperatureCode, u.reasoningRepair())
	if err != nil {
		return "", fmt.Errorf("u.complete: %w", err)
	}

	return extractHTML(raw), nil
}

func (u *UseCase) repairWithIssues(ctx context.Context, html, runtimeError string) (string, error) {
	var (
		attempts int
		current  = html
	)

	for {
		issues, err := u.validator.Check(ctx, current)
		if err != nil {
			return "", fmt.Errorf("validator.Check: %w", err)
		}

		if len(issues) == 0 && runtimeError == "" {
			return current, nil
		}

		if attempts >= u.maxFixAttempts() {
			return "", fmt.Errorf("%w: fix limit reached", domain.ErrFixLimitReached)
		}

		attempts++

		if len(issues) == 0 {
			issues = []domain.Issue{{Code: "runtime_error", Message: runtimeError}}
		}

		system, err := u.llm.Prompt("repair")
		if err != nil {
			return "", fmt.Errorf("llm.Prompt: %w", err)
		}

		p := &pipeline{}

		repaired, err := u.complete(ctx, p, system, buildRepairPrompt(current, issues), temperatureCode, u.reasoningRepair())
		if err != nil {
			return "", fmt.Errorf("u.complete: %w", err)
		}

		current = extractHTML(repaired)
		runtimeError = ""
	}
}
