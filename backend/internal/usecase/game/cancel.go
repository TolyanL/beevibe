package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

const cancelWaitInterval = 100 * time.Millisecond

func (u *UseCase) trackTask(gameID uuid.UUID, cancel context.CancelFunc, ctx context.Context) (context.Context, func()) {
	task := &runningTask{cancel: cancel}

	u.mu.Lock()
	u.running[gameID] = task
	u.mu.Unlock()

	release := func() {
		u.mu.Lock()

		if u.running[gameID] == task {
			delete(u.running, gameID)
		}

		u.mu.Unlock()
	}

	return context.WithValue(ctx, taskKey{}, task), release
}

func (u *UseCase) markReplied(ctx context.Context) {
	task, ok := ctx.Value(taskKey{}).(*runningTask)
	if !ok {
		return
	}

	task.replied.Store(true)
}

func (u *UseCase) replied(gameID uuid.UUID) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	task, ok := u.running[gameID]
	if !ok {
		return false
	}

	return task.replied.Load()
}

func (u *UseCase) Cancel(ctx context.Context, input dto.CancelGameInput) (dto.CancelGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Cancel")
	defer span.End()

	var output dto.CancelGameOutput

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	if !u.cancelRunning(input.GameID) {
		return output, domain.ErrNothingToCancel
	}

	u.waitStopped(ctx, input.GameID)

	reloaded, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	output.Status = string(reloaded.Status)

	return output, nil
}

func (u *UseCase) cancelRunning(gameID uuid.UUID) bool {
	u.mu.Lock()
	task, ok := u.running[gameID]
	u.mu.Unlock()

	if !ok {
		return false
	}

	task.cancel()

	return true
}

func (u *UseCase) waitStopped(ctx context.Context, gameID uuid.UUID) {
	deadline := time.NewTimer(u.worker.Timeout)
	defer deadline.Stop()

	ticker := time.NewTicker(cancelWaitInterval)
	defer ticker.Stop()

	for {
		if !u.isRunning(gameID) {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-deadline.C:
			return
		case <-ticker.C:
		}
	}
}

func (u *UseCase) isRunning(gameID uuid.UUID) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	_, ok := u.running[gameID]

	return ok
}
