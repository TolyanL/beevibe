package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func (u *UseCase) RunNextGameJob(ctx context.Context) (bool, error) {
	ctx, span := tracer.Start(ctx, "usecase game RunNextGameJob")
	defer span.End()

	var job domain.GameJob

	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		claimed, err := u.store.ClaimGameJob(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}

			return fmt.Errorf("store.ClaimGameJob: %w", err)
		}

		job = claimed

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if job.ID == uuid.Nil {
		return false, nil
	}

	if err := u.RunJob(ctx, job); err != nil {
		return true, fmt.Errorf("u.RunJob: %w", err)
	}

	return true, nil
}

func (u *UseCase) RunJob(ctx context.Context, job domain.GameJob) error {
	ctx, span := tracer.Start(ctx, "usecase game RunJob")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, u.worker.Timeout)
	defer cancel()

	ctx, release := u.trackTask(job.GameID, cancel, ctx)
	defer release()

	game, err := u.store.GetGame(ctx, job.GameID)
	if err != nil {
		return fmt.Errorf("store.GetGame: %w", err)
	}

	switch job.Kind {
	case domain.JobKindPatch:
		return u.runPatchJob(ctx, job, game)
	case domain.JobKindFix:
		return u.runFixJob(ctx, job, game)
	case domain.JobKindMeta:
		return u.runMetaJob(ctx, job, game)
	default:
		return fmt.Errorf("%w: unknown job kind %q", domain.ErrInvalidInput, job.Kind)
	}
}

func (u *UseCase) runPatchJob(ctx context.Context, job domain.GameJob, game domain.Game) error {
	p := &pipeline{game: game}

	latest, err := u.store.GetLatestVersion(ctx, job.GameID)
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("store.GetLatestVersion: %w", err))
	}

	history, err := u.patchHistory(ctx, job.GameID)
	if err != nil {
		return u.abortJob(ctx, p, job, err)
	}

	system, err := u.llm.Prompt("patch")
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("llm.Prompt: %w", err))
	}

	raw, err := u.complete(ctx, p, system, buildPatchPrompt(latest.HTML, job.Instruction, history), temperatureCode, u.reasoningPatch())
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("u.complete: %w", err))
	}

	patched := extractHTML(raw)

	issues, err := u.validator.Check(ctx, patched)
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("validator.Check: %w", err))
	}

	if len(issues) > 0 {
		return u.abortJob(ctx, p, job, fmt.Errorf("%w: %s", domain.ErrValidationFailed, issues[0].Message))
	}

	version, err := u.buildJobVersion(ctx, p, patched, domain.VersionSourcePatch, job.Instruction)
	if err != nil {
		return u.abortJob(ctx, p, job, err)
	}

	return u.finishJob(ctx, p, job, version)
}

func (u *UseCase) runFixJob(ctx context.Context, job domain.GameJob, game domain.Game) error {
	p := &pipeline{game: game}

	latest, err := u.store.GetLatestVersion(ctx, job.GameID)
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("store.GetLatestVersion: %w", err))
	}

	html, err := u.repairWithIssues(ctx, latest.HTML, job.Instruction)
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("u.repairWithIssues: %w", err))
	}

	version, err := u.buildJobVersion(ctx, p, html, domain.VersionSourceRepair, "repair: "+job.Instruction)
	if err != nil {
		return u.abortJob(ctx, p, job, err)
	}

	return u.finishJob(ctx, p, job, version)
}

func (u *UseCase) buildJobVersion(
	ctx context.Context,
	p *pipeline,
	html string,
	source domain.VersionSource,
	note string,
) (domain.Version, error) {
	number, err := u.nextVersionNumber(ctx, p.game.ID)
	if err != nil {
		return domain.Version{}, err
	}

	version, err := domain.NewVersion(p.game.ID, number, html, source)
	if err != nil {
		return domain.Version{}, fmt.Errorf("domain.NewVersion: %w", err)
	}

	version.Note = note
	version.TokensCost = p.tokens
	version.ModelCalls = p.calls
	version.PromptVersion = u.llm.PromptVersion()

	return version, nil
}

func (u *UseCase) finishJob(ctx context.Context, p *pipeline, job domain.GameJob, version domain.Version) error {
	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.CreateVersion(ctx, version); err != nil {
			return fmt.Errorf("store.CreateVersion: %w", err)
		}

		if err := p.game.StartStage(domain.StatusChecking); err != nil {
			return fmt.Errorf("game.StartStage: %w", err)
		}

		if err := u.store.UpdateGameStatus(ctx, p.game.ID, p.game.Status, ""); err != nil {
			return fmt.Errorf("store.UpdateGameStatus: %w", err)
		}

		if err := u.publishEvent(ctx, p.game, domain.EventGamePatched, &version.ID, p.tokens, ""); err != nil {
			return fmt.Errorf("u.publishEvent: %w", err)
		}

		return nil
	})
	if err != nil {
		return u.abortJob(ctx, p, job, fmt.Errorf("transaction.Wrap: %w", err))
	}

	if err := u.emitOperationStageEvent(ctx, p.game.ID, job.Kind, "done"); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	if err := u.transition(ctx, p, domain.StatusReady); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusChecking, "done", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	if err := u.emitArtifactEvent(ctx, p.game.ID, "version", map[string]any{
		"versionId": version.ID.String(),
		"number":    version.Number,
		"source":    string(version.Source),
	}); err != nil {
		return fmt.Errorf("u.emitArtifactEvent: %w", err)
	}

	if err := u.emitDoneEvent(ctx, p.game.ID, version.ID, p.tokens, p.calls); err != nil {
		return fmt.Errorf("u.emitDoneEvent: %w", err)
	}

	return nil
}

func (u *UseCase) abortJob(ctx context.Context, p *pipeline, job domain.GameJob, cause error) error {
	failCtx := context.WithoutCancel(ctx)

	err := transaction.Wrap(failCtx, func(ctx context.Context) error {
		if job.Kind == domain.JobKindPatch && !u.replied(job.GameID) {
			if err := u.refundInTx(ctx, job.OwnerID, job.GameID, u.tokens.PricePatch); err != nil {
				return fmt.Errorf("u.refundInTx: %w", err)
			}
		}

		if err := p.game.StartStage(domain.StatusReady); err != nil {
			return fmt.Errorf("game.StartStage: %w", err)
		}

		if err := u.store.UpdateGameStatus(ctx, p.game.ID, p.game.Status, ""); err != nil {
			return fmt.Errorf("store.UpdateGameStatus: %w", err)
		}

		return nil
	})
	if err != nil {
		return errors.Join(cause, fmt.Errorf("transaction.Wrap: %w", err))
	}

	if err := u.emitOperationStageEvent(failCtx, p.game.ID, job.Kind, "error"); err != nil {
		return errors.Join(cause, fmt.Errorf("u.emitStageEvent: %w", err))
	}

	if err := u.emitErrorEvent(failCtx, p.game.ID, errorCode(cause), cause.Error()); err != nil {
		return errors.Join(cause, fmt.Errorf("u.emitErrorEvent: %w", err))
	}

	return cause
}

func (u *UseCase) markPatching(ctx context.Context, game *domain.Game, kind domain.JobKind) error {
	if err := game.StartStage(domain.StatusPatching); err != nil {
		return fmt.Errorf("game.StartStage: %w", err)
	}

	if err := u.store.UpdateGameStatus(ctx, game.ID, game.Status, ""); err != nil {
		return fmt.Errorf("store.UpdateGameStatus: %w", err)
	}

	return u.emitOperationStageEvent(ctx, game.ID, kind, "running")
}
