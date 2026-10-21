package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) RecordPlaytime(ctx context.Context, input dto.RecordPlaytimeInput) (dto.RecordPlaytimeOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase RecordPlaytime")
	defer span.End()

	var output dto.RecordPlaytimeOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	seconds := domain.ClampPlaytime(input.Seconds)
	userID := input.UserID

	session := domain.PlaySession{
		ID:      uuid.New(),
		GameID:  input.GameID,
		UserID:  &userID,
		Seconds: seconds,
	}

	if err := u.store.InsertPlaySession(ctx, session); err != nil {
		return output, fmt.Errorf("store.InsertPlaySession: %w", err)
	}

	aggregate, err := u.store.GetPlaytimeAggregate(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetPlaytimeAggregate: %w", err)
	}

	output.TotalSeconds = aggregate.Seconds

	return output, nil
}
