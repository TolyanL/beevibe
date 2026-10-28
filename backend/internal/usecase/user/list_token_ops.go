package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) ListTokenOps(ctx context.Context, userID uuid.UUID, input dto.ListTokenOpsInput) (dto.ListTokenOpsOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase user ListTokenOps")
	defer span.End()

	var output dto.ListTokenOpsOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 20
	}

	ops, total, err := u.store.ListTokenOps(ctx, userID, dto.ListTokenOpsInput{Limit: limit, Offset: input.Offset})
	if err != nil {
		return output, fmt.Errorf("store.ListTokenOps: %w", err)
	}

	output.Items = make([]dto.ProfileTokenOp, 0, len(ops))
	output.Total = total

	for _, op := range ops {
		output.Items = append(output.Items, dto.ProfileTokenOp{
			GameID:    op.GameID,
			Delta:     op.Delta,
			Reason:    string(op.Reason),
			CreatedAt: op.CreatedAt,
		})
	}

	return output, nil
}
