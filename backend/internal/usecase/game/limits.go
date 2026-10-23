package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/llm"
)

func errorCode(err error) string {
	switch {
	case errors.Is(err, domain.ErrLLMUnavailable):
		return "llm_unavailable"
	case errors.Is(err, domain.ErrLLMBudgetExceeded):
		return "llm_budget_exceeded"
	case errors.Is(err, domain.ErrValidationFailed):
		return "validation_failed"
	case errors.Is(err, domain.ErrNotEnoughTokens):
		return "not_enough_tokens"
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	default:
		return "generation_failed"
	}
}

func (u *UseCase) reasoningAnalyze() llm.Effort {
	return u.reasoning.Analyze
}

func (u *UseCase) reasoningPlan() llm.Effort {
	return u.reasoning.Plan
}

func (u *UseCase) reasoningCode() llm.Effort {
	return u.reasoning.Code
}

func (u *UseCase) reasoningAssets() llm.Effort {
	return u.reasoning.Assets
}

func (u *UseCase) reasoningReview() llm.Effort {
	return u.reasoning.Review
}

func (u *UseCase) reasoningAssemble() llm.Effort {
	return u.reasoning.Assemble
}

func (u *UseCase) reasoningPatch() llm.Effort {
	return u.reasoning.Patch
}

func (u *UseCase) reasoningRepair() llm.Effort {
	return u.reasoning.Repair
}

func (u *UseCase) maxFixAttempts() int {
	if u.worker.MaxFixAttempts < 1 {
		return 1
	}

	return u.worker.MaxFixAttempts
}

func joinIssues(issues []domain.Issue) string {
	out := ""

	for _, issue := range issues {
		out += fmt.Sprintf("[%s] %s\n", issue.Code, issue.Message)
	}

	return out
}
