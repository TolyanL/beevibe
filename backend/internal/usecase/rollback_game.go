package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) RollbackGame(ctx context.Context, input dto.RollbackGameInput) (dto.RollbackGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase RollbackGame")
	defer span.End()

	var output dto.RollbackGameOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	if game.Status != domain.StatusReady {
		return output, domain.ErrGenerationInProgress
	}

	source, err := u.store.GetVersion(ctx, input.GameID, input.Version)
	if err != nil {
		return output, fmt.Errorf("store.GetVersion: %w", err)
	}

	number, err := u.nextVersionNumber(ctx, input.GameID)
	if err != nil {
		return output, err
	}

	version, err := domain.NewVersion(input.GameID, number, source.HTML, domain.VersionSourceRollback)
	if err != nil {
		return output, fmt.Errorf("domain.NewVersion: %w", err)
	}

	version.Note = fmt.Sprintf("rollback to v%d", source.Number)

	if err := u.store.CreateVersion(ctx, version); err != nil {
		return output, fmt.Errorf("store.CreateVersion: %w", err)
	}

	output.VersionID = version.ID

	return output, nil
}
