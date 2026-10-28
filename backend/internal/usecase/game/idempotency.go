package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

const (
	operationCreateGame = "create_game"
	operationPatchGame  = "patch_game"
	operationFixGame    = "fix_game"
)

func (u *UseCase) claimIdempotencyKey(ctx context.Context, key, operation string, userID uuid.UUID) (string, bool, error) {
	inserted, err := u.store.InsertIdempotencyKey(ctx, key, operation, userID)
	if err != nil {
		return "", false, fmt.Errorf("store.InsertIdempotencyKey: %w", err)
	}

	if inserted {
		return "", false, nil
	}

	resultID, found, err := u.store.GetIdempotencyResult(ctx, key, operation)
	if err != nil {
		return "", false, fmt.Errorf("store.GetIdempotencyResult: %w", err)
	}

	if !found || resultID == "" {
		return "", false, domain.ErrIdempotencyInProgress
	}

	return resultID, true, nil
}

func (u *UseCase) completeIdempotencyKey(ctx context.Context, key, operation, resultID string) error {
	if key == "" {
		return nil
	}

	if err := u.store.CompleteIdempotencyKey(ctx, key, operation, resultID); err != nil {
		return fmt.Errorf("store.CompleteIdempotencyKey: %w", err)
	}

	return nil
}

func (u *UseCase) releaseIdempotencyKey(ctx context.Context, key, operation string) {
	if key == "" {
		return
	}

	if err := u.store.DeleteIdempotencyKey(ctx, key, operation); err != nil {
		return
	}
}
