package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) GetGamePublic(ctx context.Context, input dto.GetGamePublicInput) (dto.GetGamePublicOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase GetGamePublic")
	defer span.End()

	var output dto.GetGamePublicOutput

	version, err := u.store.GetLatestVersion(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetLatestVersion: %w", err)
	}

	output.HTML = version.HTML

	return output, nil
}
