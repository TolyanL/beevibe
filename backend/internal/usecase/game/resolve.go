package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

// ResolveGameID переводит slug из адреса в идентификатор игры: наружу
// показываем только читаемые ссылки, uuid остаётся внутри API.
func (u *UseCase) ResolveGameID(ctx context.Context, slug string) (uuid.UUID, error) {
	ctx, span := tracer.Start(ctx, "usecase game ResolveGameID")
	defer span.End()

	if slug == "" {
		return uuid.Nil, fmt.Errorf("%w: empty game slug", domain.ErrInvalidInput)
	}

	game, err := u.store.GetGameBySlug(ctx, slug)
	if err != nil {
		return uuid.Nil, fmt.Errorf("store.GetGameBySlug: %w", err)
	}

	return game.ID, nil
}
