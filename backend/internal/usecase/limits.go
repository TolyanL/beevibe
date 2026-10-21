package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
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

func (u *UseCase) maxTokensAnalyze() int {
	return u.limits.MaxTokensAnalyze
}

func (u *UseCase) maxTokensPlan() int {
	return u.limits.MaxTokensPlan
}

func (u *UseCase) maxTokensCode() int {
	return u.limits.MaxTokensCode
}

func (u *UseCase) maxTokensAssets() int {
	return u.limits.MaxTokensAssets
}

func (u *UseCase) maxTokensReview() int {
	return u.limits.MaxTokensReview
}

func (u *UseCase) maxTokensAssemble() int {
	return u.limits.MaxTokensAssemble
}

func (u *UseCase) maxTokensPatch() int {
	return u.limits.MaxTokensPatch
}

func (u *UseCase) maxTokensRepair() int {
	return u.limits.MaxTokensRepair
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
