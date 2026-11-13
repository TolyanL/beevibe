package game

import (
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/llm"
)

func (u *UseCase) schedule(p *pipeline) []stageStep {
	if p.fast() {
		return []stageStep{
			{domain.StatusAnalyzing, u.stepSpecPlan},
			{domain.StatusCoding, u.stepOneShot},
			{domain.StatusChecking, u.stepCheck},
		}
	}

	return []stageStep{
		{domain.StatusAnalyzing, u.stepAnalyze},
		{domain.StatusPlanning, u.stepPlan},
		{domain.StatusCoding, u.stepCode},
		{domain.StatusAssets, u.stepAssets},
		{domain.StatusReviewing, u.stepReview},
		{domain.StatusAssembling, u.stepAssemble},
		{domain.StatusChecking, u.stepCheck},
	}
}

// NOTE: the fast schedule is a subsequence of the full pipeline order, so a resume stage
// produced by the full order has to be folded into the nearest fast stage.
func resumeStageFor(p *pipeline, resume domain.Status) domain.Status {
	if !p.fast() {
		return resume
	}

	switch resume {
	case domain.StatusAnalyzing:
		return domain.StatusAnalyzing
	case domain.StatusPlanning, domain.StatusCoding:
		return domain.StatusCoding
	default:
		return domain.StatusChecking
	}
}

func (u *UseCase) timeout(p *pipeline) time.Duration {
	if p.fast() && u.fast.Timeout > 0 {
		return u.fast.Timeout
	}

	return u.worker.Timeout
}

func (u *UseCase) client(p *pipeline) llm.LLM {
	if p.fast() && u.fast.LLM != nil {
		return u.fast.LLM
	}

	return u.llm
}

func (u *UseCase) maxCalls(p *pipeline) int {
	if p.fast() && u.fast.MaxCallsPerGame > 0 {
		return u.fast.MaxCallsPerGame
	}

	return u.worker.MaxCallsPerGame
}

func (u *UseCase) maxFixAttemptsFor(p *pipeline) int {
	if p.fast() && u.fast.MaxFixAttempts > 0 {
		return u.fast.MaxFixAttempts
	}

	return u.maxFixAttempts()
}

func (u *UseCase) generatePrice(speed domain.Speed) int64 {
	if speed.IsFast() && u.tokens.PriceGenerateFast > 0 {
		return u.tokens.PriceGenerateFast
	}

	return u.tokens.PriceGenerate
}
