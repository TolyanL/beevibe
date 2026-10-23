package usecase

import (
	"context"

	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) Health(ctx context.Context) (dto.HealthOutput, error) {
	_, span := tracer.Start(ctx, "usecase Health")
	defer span.End()

	return dto.HealthOutput{Status: "ok", Version: u.app.Version}, nil
}
