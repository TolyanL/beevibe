package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

// ResolveGameID принимает uuid или slug и возвращает идентификатор игры.
func (u *UseCase) ResolveGameID(ctx context.Context, ref string) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "usecase game ResolveGameID")
	defer span.End()

	if gameID, err := uuid.Parse(ref); err == nil {
		return gameID, nil
	}

	if ref == "" {
		return uuid.Nil, fmt.Errorf("%w: empty game reference", domain.ErrInvalidInput)
	}

	game, err := u.store.GetGameBySlug(ctx, ref)
	if err != nil {
		return uuid.Nil, fmt.Errorf("store.GetGameBySlug: %w", err)
	}

	return game.ID, nil
}
