package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/llm"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const (
	temperatureCode   = 0.2
	temperatureIdeas  = 0.7
	temperatureReview = 0.3
)

type pipeline struct {
	game   domain.Game
	stage  domain.Status
	spec   domain.GameSpec
	plan   string
	code   string
	assets string
	html   string
	tokens int
	calls  int
}

func (u *UseCase) RunNextQueuedGame(ctx context.Context) (bool, error) {
	ctx, span := tracer.Start(ctx, "usecase game RunNextQueuedGame")
	defer span.End()

	var game domain.Game

	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		claimed, err := u.store.ClaimQueuedGame(ctx)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return nil
			}

			return fmt.Errorf("store.ClaimQueuedGame: %w", err)
		}

		game = claimed

		return nil
	})
	if err != nil {
		return false, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if game.ID == uuid.Nil {
		return false, nil
	}

	if err := u.RunPipeline(ctx, game); err != nil {
		return true, fmt.Errorf("u.RunPipeline: %w", err)
	}

	return true, nil
}

func (u *UseCase) RunPipeline(ctx context.Context, game domain.Game) error {
	ctx, span := tracer.Start(ctx, "usecase game RunPipeline")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, u.worker.Timeout)
	defer cancel()

	ctx, release := u.trackTask(game.ID, cancel, ctx)
	defer release()

	p := &pipeline{game: game}

	resume, err := u.restorePipeline(ctx, p)
	if err != nil {
		return u.abort(ctx, p, err)
	}

	if resume != domain.StatusAnalyzing {
		if err := p.game.ResumeAt(resume); err != nil {
			return u.abort(ctx, p, fmt.Errorf("game.ResumeAt: %w", err))
		}

		if err := u.store.UpdateGameStatus(ctx, p.game.ID, p.game.Status, ""); err != nil {
			return u.abort(ctx, p, fmt.Errorf("store.UpdateGameStatus: %w", err))
		}
	}

	steps := []struct {
		stage domain.Status
		run   func(context.Context, *pipeline) error
	}{
		{domain.StatusAnalyzing, u.stepAnalyze},
		{domain.StatusPlanning, u.stepPlan},
		{domain.StatusCoding, u.stepCode},
		{domain.StatusAssets, u.stepAssets},
		{domain.StatusReviewing, u.stepReview},
		{domain.StatusAssembling, u.stepAssemble},
		{domain.StatusChecking, u.stepCheck},
	}

	for _, step := range steps {
		if step.stage.Before(resume) {
			continue
		}

		if err := step.run(ctx, p); err != nil {
			if errors.Is(err, errNeedsDecision) {
				return nil
			}

			return u.abort(ctx, p, err)
		}
	}

	return nil
}

func (u *UseCase) restorePipeline(ctx context.Context, p *pipeline) (domain.Status, error) {
	events, err := u.store.ListGenEvents(ctx, p.game.ID, 0)
	if err != nil {
		return domain.StatusAnalyzing, fmt.Errorf("store.ListGenEvents: %w", err)
	}

	artifacts, err := u.store.ListGameArtifacts(ctx, p.game.ID)
	if err != nil {
		return domain.StatusAnalyzing, fmt.Errorf("store.ListGameArtifacts: %w", err)
	}

	p.plan = artifacts[artifactPlan]
	p.code = artifacts[artifactCode]
	p.assets = artifacts[artifactAssets]
	p.html = artifacts[artifactHTML]

	if p.game.Spec != nil && p.spec.Genre == "" {
		p.spec = *p.game.Spec
	}

	if p.game.ResumeStage != "" {
		hint := p.game.ResumeStage

		if err := u.store.SaveGameResumeStage(ctx, p.game.ID, ""); err != nil {
			return domain.StatusAnalyzing, fmt.Errorf("store.SaveGameResumeStage: %w", err)
		}

		return hint, nil
	}

	return resumeStage(events), nil
}

func (u *UseCase) stepAnalyze(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusAnalyzing, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("analyzer")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	prompt := buildAnalyzePrompt(p.game)

	raw, err := u.complete(ctx, p, system, prompt, temperatureIdeas, u.reasoningAnalyze())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	payload, err := extractJSONObject(raw)
	if err != nil {
		return fmt.Errorf("extractJSONObject: %w", err)
	}

	var spec domain.GameSpec

	if err := json.Unmarshal([]byte(payload), &spec); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	p.spec = spec

	if err := p.game.ApplySpec(spec); err != nil {
		return fmt.Errorf("game.ApplySpec: %w", err)
	}

	if err := u.applyCoverFromSpec(ctx, &p.game); err != nil {
		return fmt.Errorf("u.applyCoverFromSpec: %w", err)
	}

	if err := u.store.UpdateGame(ctx, p.game); err != nil {
		return fmt.Errorf("store.UpdateGame: %w", err)
	}

	if err := u.emitArtifactEvent(ctx, p.game.ID, "spec", spec); err != nil {
		return fmt.Errorf("u.emitArtifactEvent: %w", err)
	}

	return u.emitStageEvent(ctx, p.game.ID, domain.StatusAnalyzing, "done", time.Since(start).Milliseconds(), 0)
}

func (u *UseCase) stepPlan(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusPlanning); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusPlanning, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("planner")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	spec, err := marshalSpec(p.spec)
	if err != nil {
		return fmt.Errorf("marshalSpec: %w", err)
	}

	raw, err := u.complete(ctx, p, system, spec, temperatureReview, u.reasoningPlan())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	p.plan = raw

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactPlan, raw); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	return u.emitStageEvent(ctx, p.game.ID, domain.StatusPlanning, "done", time.Since(start).Milliseconds(), 0)
}

func (u *UseCase) stepCode(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusCoding); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusCoding, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("coder")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.complete(ctx, p, system, buildCodePrompt(p), temperatureCode, u.reasoningCode())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	p.code = raw

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactCode, raw); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	return u.emitStageEvent(ctx, p.game.ID, domain.StatusCoding, "done", time.Since(start).Milliseconds(), 0)
}

func (u *UseCase) stepAssets(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusAssets); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusAssets, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("assets")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	spec, err := marshalSpec(p.spec)
	if err != nil {
		return fmt.Errorf("marshalSpec: %w", err)
	}

	raw, err := u.complete(ctx, p, system, spec, temperatureIdeas, u.reasoningAssets())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	p.assets = raw

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactAssets, raw); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	return u.emitStageEvent(ctx, p.game.ID, domain.StatusAssets, "done", time.Since(start).Milliseconds(), 0)
}

func (u *UseCase) stepReview(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusReviewing); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusReviewing, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("reviewer")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.complete(ctx, p, system, buildReviewPrompt(p), temperatureReview, u.reasoningReview())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	payload, err := extractJSONObject(raw)
	if err != nil {
		return fmt.Errorf("extractJSONObject: %w", err)
	}

	var review struct {
		Issues []domain.Issue `json:"issues"`
	}

	if err := json.Unmarshal([]byte(payload), &review); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	duration := time.Since(start).Milliseconds()

	if len(review.Issues) == 0 {
		return u.emitStageEvent(ctx, p.game.ID, domain.StatusReviewing, "done", duration, 0)
	}

	issuesRaw, err := json.Marshal(review.Issues)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactReviewIssues, string(issuesRaw)); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	if err := u.transition(ctx, p, domain.StatusNeedsDecision); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusReviewing, "done", duration, len(review.Issues)); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	return errNeedsDecision
}

func (u *UseCase) stepAssemble(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusAssembling); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusAssembling, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("assembler")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.complete(ctx, p, system, buildAssemblePrompt(p), temperatureCode, u.reasoningAssemble())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	html := extractHTML(raw)

	for attempt := range u.worker.MaxFixAttempts {
		issues, err := u.validator.Check(ctx, html)
		if err != nil {
			return fmt.Errorf("validator.Check: %w", err)
		}

		if len(issues) == 0 {
			p.html = html

			if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactHTML, html); err != nil {
				return fmt.Errorf("store.SaveGameArtifact: %w", err)
			}

			return u.emitStageEvent(ctx, p.game.ID, domain.StatusAssembling, "done", time.Since(start).Milliseconds(), 0)
		}

		if attempt == u.worker.MaxFixAttempts-1 {
			return fmt.Errorf("%w: %s", domain.ErrValidationFailed, issues[0].Message)
		}

		repaired, err := u.repair(ctx, p, html, issues)
		if err != nil {
			return fmt.Errorf("u.repair: %w", err)
		}

		html = repaired
	}

	return fmt.Errorf("%w: validation issues remain", domain.ErrValidationFailed)
}

func (u *UseCase) stepCheck(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusChecking); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusChecking, "running", 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	cfg, err := u.validator.ExtractConfig(ctx, p.html)
	if err != nil {
		return fmt.Errorf("validator.ExtractConfig: %w", err)
	}

	if err := p.game.ApplyConfig(cfg); err != nil {
		return fmt.Errorf("game.ApplyConfig: %w", err)
	}

	if err := u.store.UpdateGame(ctx, p.game); err != nil {
		return fmt.Errorf("store.UpdateGame: %w", err)
	}

	if err := u.emitArtifactEvent(ctx, p.game.ID, "config", cfg); err != nil {
		return fmt.Errorf("u.emitArtifactEvent: %w", err)
	}

	version, err := domain.NewVersion(p.game.ID, 1, p.html, domain.VersionSourceGenerate)
	if err != nil {
		return fmt.Errorf("domain.NewVersion: %w", err)
	}

	version.TokensCost = p.tokens
	version.ModelCalls = p.calls
	version.PromptVersion = u.llm.PromptVersion()

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.CreateVersion(ctx, version); err != nil {
			return fmt.Errorf("store.CreateVersion: %w", err)
		}

		if err := u.publishEvent(ctx, p.game, domain.EventGameReady, &version.ID, p.tokens, ""); err != nil {
			return fmt.Errorf("u.publishEvent: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction.Wrap: %w", err)
	}

	if err := u.transition(ctx, p, domain.StatusReady); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusChecking, "done", time.Since(start).Milliseconds(), 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	if err := u.emitDoneEvent(ctx, p.game.ID, version.ID, p.tokens, p.calls); err != nil {
		return fmt.Errorf("u.emitDoneEvent: %w", err)
	}

	return nil
}

func (u *UseCase) complete(
	ctx context.Context,
	p *pipeline,
	system, prompt string,
	temperature float64,
	effort llm.Effort,
) (string, error) {
	if p.calls >= u.worker.MaxCallsPerGame {
		return "", domain.ErrLLMBudgetExceeded
	}

	resp, err := u.llm.Complete(ctx, llm.Request{
		System:      system,
		Prompt:      prompt,
		Temperature: temperature,
		Reasoning:   effort,
	})
	if err != nil {
		return "", fmt.Errorf("llm.Complete: %w", err)
	}

	p.calls++
	p.tokens += resp.TotalTokens()
	u.markReplied(ctx)

	return resp.Text, nil
}

func (u *UseCase) transition(ctx context.Context, p *pipeline, stage domain.Status) error {
	p.stage = stage

	if p.game.Status == stage {
		return nil
	}

	if err := p.game.StartStage(stage); err != nil {
		return fmt.Errorf("game.StartStage: %w", err)
	}

	if err := u.store.UpdateGameStatus(ctx, p.game.ID, p.game.Status, p.game.FailReason); err != nil {
		return fmt.Errorf("store.UpdateGameStatus: %w", err)
	}

	return nil
}

func (u *UseCase) abort(ctx context.Context, p *pipeline, cause error) error {
	failCtx := context.WithoutCancel(ctx)

	stage := p.stage
	if stage == "" {
		stage = p.game.Status
	}

	if stage.IsPipeline() {
		if err := u.emitStageEvent(failCtx, p.game.ID, stage, "error", 0, 0); err != nil {
			cause = errors.Join(cause, fmt.Errorf("u.emitStageEvent: %w", err))
		}
	}

	if err := u.failGame(failCtx, p, cause); err != nil {
		return errors.Join(cause, err)
	}

	return cause
}

func (u *UseCase) failGame(ctx context.Context, p *pipeline, cause error) error {
	reason := cause.Error()

	err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := p.game.Fail(reason); err != nil {
			return fmt.Errorf("game.Fail: %w", err)
		}

		if err := u.store.UpdateGameStatus(ctx, p.game.ID, p.game.Status, p.game.FailReason); err != nil {
			return fmt.Errorf("store.UpdateGameStatus: %w", err)
		}

		if !u.replied(p.game.ID) {
			if err := u.refundInTx(ctx, p.game.OwnerID, p.game.ID, u.tokens.PriceGenerate); err != nil {
				return fmt.Errorf("u.refundInTx: %w", err)
			}
		}

		if err := u.publishEvent(ctx, p.game, domain.EventGameFailed, nil, p.tokens, reason); err != nil {
			return fmt.Errorf("u.publishEvent: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction.Wrap: %w", err)
	}

	if err := u.emitErrorEvent(ctx, p.game.ID, errorCode(cause), reason); err != nil {
		return fmt.Errorf("u.emitErrorEvent: %w", err)
	}

	return nil
}
