package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) Fix(ctx context.Context, input dto.FixGameInput) (dto.FixGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Fix")
	defer span.End()

	var output dto.FixGameOutput

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

	html, err := u.repairWithIssues(ctx, latest.HTML, input.Error)
	if err != nil {
		return output, err
	}

	number, err := u.nextVersionNumber(ctx, input.GameID)
	if err != nil {
		return output, err
	}

	version, err := domain.NewVersion(input.GameID, number, html, domain.VersionSourceRepair)
	if err != nil {
		return output, fmt.Errorf("domain.NewVersion: %w", err)
	}

	version.Note = "repair: " + input.Error

	if err := u.store.CreateVersion(ctx, version); err != nil {
		return output, fmt.Errorf("store.CreateVersion: %w", err)
	}

	output.VersionID = version.ID

	return output, nil
}

func (u *UseCase) repair(ctx context.Context, p *pipeline, html string, issues []domain.Issue) (string, error) {
	system, err := u.llm.Prompt("repair")
	if err != nil {
		return "", fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.complete(ctx, p, system, buildRepairPrompt(html, issues), temperatureCode, u.maxTokensRepair())
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

		repaired, err := u.complete(ctx, p, system, buildRepairPrompt(current, issues), temperatureCode, u.maxTokensRepair())
		if err != nil {
			return "", fmt.Errorf("u.complete: %w", err)
		}

		current = extractHTML(repaired)
		runtimeError = ""
	}
}
