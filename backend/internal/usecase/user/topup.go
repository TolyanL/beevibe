package user

import (
	"context"
	"fmt"
	"time"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const (
	operationTopupTokens = "topup_tokens"
	topupWindow          = 24 * time.Hour
)

func (u *UseCase) TopupTokens(ctx context.Context, input dto.TopupTokensInput) (dto.TopupTokensOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase user TopupTokens")
	defer span.End()

	var output dto.TopupTokensOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	if !u.tokens.DemoTopup {
		return output, domain.ErrTopupDisabled
	}

	if err := u.validateTopupAmount(input.Amount); err != nil {
		return output, err
	}

	if input.IdempotencyKey != "" {
		replay, err := u.replayTopup(ctx, input)
		if err != nil {
			return output, err
		}

		if replay != nil {
			return *replay, nil
		}
	}

	used, err := u.store.SumTopupsSince(ctx, input.UserID, time.Now().UTC().Add(-topupWindow))
	if err != nil {
		return output, fmt.Errorf("store.SumTopupsSince: %w", err)
	}

	if used+input.Amount > u.tokens.TopupDailyCap {
		return output, domain.ErrTopupLimitReached
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		user, err := u.store.GetUserByIDForUpdate(ctx, input.UserID)
		if err != nil {
			return fmt.Errorf("store.GetUserByIDForUpdate: %w", err)
		}

		if err := user.Credit(input.Amount); err != nil {
			return fmt.Errorf("user.Credit: %w", err)
		}

		if err := u.store.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("store.UpdateUser: %w", err)
		}

		op, err := domain.NewTokenOp(user.ID, nil, input.Amount, domain.ReasonTopup)
		if err != nil {
			return fmt.Errorf("domain.NewTokenOp: %w", err)
		}

		if err := u.store.InsertTokenOp(ctx, op); err != nil {
			return fmt.Errorf("store.InsertTokenOp: %w", err)
		}

		output.Balance = user.TokenBalance

		return nil
	})
	if err != nil {
		u.releaseTopupKey(ctx, input)

		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if input.IdempotencyKey != "" {
		if err := u.store.CompleteIdempotencyKey(ctx, input.IdempotencyKey, operationTopupTokens, fmt.Sprintf("%d", output.Balance)); err != nil {
			return output, fmt.Errorf("store.CompleteIdempotencyKey: %w", err)
		}
	}

	return output, nil
}

func (u *UseCase) validateTopupAmount(amount int64) error {
	minAmount := u.tokens.TopupMin
	if minAmount <= 0 {
		minAmount = 1
	}

	maxAmount := u.tokens.TopupMax
	if maxAmount <= 0 {
		maxAmount = minAmount
	}

	if amount < minAmount || amount > maxAmount {
		return fmt.Errorf("%w: amount must be between %d and %d", domain.ErrInvalidInput, minAmount, maxAmount)
	}

	return nil
}

func (u *UseCase) replayTopup(ctx context.Context, input dto.TopupTokensInput) (*dto.TopupTokensOutput, error) {
	inserted, err := u.store.InsertIdempotencyKey(ctx, input.IdempotencyKey, operationTopupTokens, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("store.InsertIdempotencyKey: %w", err)
	}

	if inserted {
		return nil, nil
	}

	result, found, err := u.store.GetIdempotencyResult(ctx, input.IdempotencyKey, operationTopupTokens)
	if err != nil {
		return nil, fmt.Errorf("store.GetIdempotencyResult: %w", err)
	}

	if !found || result == "" {
		return nil, domain.ErrIdempotencyInProgress
	}

	var balance int64

	if _, err := fmt.Sscanf(result, "%d", &balance); err != nil {
		return nil, fmt.Errorf("fmt.Sscanf: %w", err)
	}

	return &dto.TopupTokensOutput{Balance: balance}, nil
}

func (u *UseCase) releaseTopupKey(ctx context.Context, input dto.TopupTokensInput) {
	if input.IdempotencyKey == "" {
		return
	}

	if err := u.store.DeleteIdempotencyKey(ctx, input.IdempotencyKey, operationTopupTokens); err != nil {
		return
	}
}
