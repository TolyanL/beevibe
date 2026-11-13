package game

import (
	"context"
	"fmt"
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
)

func (u *UseCase) stepSpecPlan(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusAnalyzing, stageStatusRunning, 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("specplan")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.complete(ctx, p, system, buildSpecPlanPrompt(p.game), temperatureIdeas, u.reasoningAnalyze())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	spec, plan, err := parseSpecPlan(raw)
	if err != nil {
		return fmt.Errorf("parseSpecPlan: %w", err)
	}

	p.spec = spec
	p.plan = plan

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

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactPlan, plan); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	return u.emitStageEvent(ctx, p.game.ID, domain.StatusAnalyzing, stageStatusDone, time.Since(start).Milliseconds(), 0)
}

func (u *UseCase) stepOneShot(ctx context.Context, p *pipeline) error {
	start := time.Now()

	if err := u.transition(ctx, p, domain.StatusCoding); err != nil {
		return fmt.Errorf("u.transition: %w", err)
	}

	if err := u.emitStageEvent(ctx, p.game.ID, domain.StatusCoding, stageStatusRunning, 0, 0); err != nil {
		return fmt.Errorf("u.emitStageEvent: %w", err)
	}

	system, err := u.llm.Prompt("onepass")
	if err != nil {
		return fmt.Errorf("llm.Prompt: %w", err)
	}

	raw, err := u.completeStream(ctx, p, system, buildOneShotPrompt(p), temperatureCode, u.reasoningCode())
	if err != nil {
		return fmt.Errorf("u.complete: %w", err)
	}

	html, err := u.fixUntilClean(ctx, p, extractHTML(raw))
	if err != nil {
		return fmt.Errorf("u.fixUntilClean: %w", err)
	}

	p.html = html

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactCode, html); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	if err := u.store.SaveGameArtifact(ctx, p.game.ID, artifactHTML, html); err != nil {
		return fmt.Errorf("store.SaveGameArtifact: %w", err)
	}

	return u.emitStageEvent(ctx, p.game.ID, domain.StatusCoding, stageStatusDone, time.Since(start).Milliseconds(), 0)
}
