package game

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) GetPublic(ctx context.Context, input dto.GetGamePublicInput) (dto.GetGamePublicOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game GetPublic")
	defer span.End()

	var output dto.GetGamePublicOutput

	game, err := u.gameBySlug(ctx, input.Slug)
	if err != nil {
		return output, err
	}

	if !visibleTo(game, input.ViewerID) {
		return output, domain.ErrNotFound
	}

	version, err := u.store.GetLatestVersion(ctx, game.ID)
	if err != nil {
		return output, fmt.Errorf("store.GetLatestVersion: %w", err)
	}

	output.HTML = version.HTML

	return output, nil
}

func (u *UseCase) GetPublicCard(ctx context.Context, input dto.GetGamePublicCardInput) (dto.GetGamePublicCardOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game GetPublicCard")
	defer span.End()

	var output dto.GetGamePublicCardOutput

	game, err := u.gameBySlug(ctx, input.Slug)
	if err != nil {
		return output, err
	}

	rating, err := u.store.GetRatingAggregate(ctx, game.ID)
	if err != nil {
		return output, fmt.Errorf("store.GetRatingAggregate: %w", err)
	}

	playtime, err := u.store.GetPlaytimeAggregate(ctx, game.ID)
	if err != nil {
		return output, fmt.Errorf("store.GetPlaytimeAggregate: %w", err)
	}

	owner, err := u.store.GetUserByID(ctx, game.OwnerID)
	if err != nil {
		return output, fmt.Errorf("store.GetUserByID: %w", err)
	}

	output.ID = game.ID
	output.Prompt = game.Prompt
	output.Description = game.Description
	output.Title = game.Title
	output.Slug = game.DisplaySlug()
	output.Author = owner.PublicName()
	output.Cover = coverOf(game)
	output.Genre = string(game.Params.Genre)
	output.Style = string(game.Params.Style)
	output.CustomGenre = string(game.Params.CustomGenre)
	output.CustomStyle = string(game.Params.CustomStyle)
	output.MapSize = string(game.Params.MapSize)
	output.CreatedAt = game.CreatedAt
	output.Rating = rating
	output.Playtime = playtime

	if input.ViewerID == nil {
		return output, nil
	}

	output.IsOwner = game.IsOwner(*input.ViewerID)

	myRating, err := u.store.GetRating(ctx, game.ID, *input.ViewerID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return output, fmt.Errorf("store.GetRating: %w", err)
	}

	if err == nil {
		output.MyRating = int(myRating.Value)
	}

	return output, nil
}

func (u *UseCase) gameBySlug(ctx context.Context, slug string) (domain.Game, error) {
	if slug == "" {
		return domain.Game{}, fmt.Errorf("%w: empty game slug", domain.ErrInvalidInput)
	}

	game, err := u.store.GetGameBySlug(ctx, slug)
	if err != nil {
		return domain.Game{}, fmt.Errorf("store.GetGameBySlug: %w", err)
	}

	return game, nil
}

// visibleTo пускает автора к своей игре в любом статусе: превью в рабочем
// пространстве не должно зависеть от публикации.
func visibleTo(game domain.Game, viewerID *uuid.UUID) bool {
	if game.IsPublic() {
		return true
	}

	return viewerID != nil && game.IsOwner(*viewerID)
}
