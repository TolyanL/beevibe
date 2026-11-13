package game

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) ListUserGenerations(ctx context.Context, input dto.ListUserGenerationsInput) (dto.ListUserGenerationsOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game ListUserGenerations")
	defer span.End()

	var output dto.ListUserGenerationsOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	items, total, err := u.store.ListUserGenerations(ctx, input.UserID, dto.ListUserGenerationsInput{
		Limit:  limit,
		Offset: input.Offset,
	})
	if err != nil {
		return output, fmt.Errorf("store.ListUserGenerations: %w", err)
	}

	output.Items = items
	output.Total = total

	return output, nil
}
