package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) Get(ctx context.Context, input dto.GetGameInput) (dto.GetGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Get")
	defer span.End()

	var output dto.GetGameOutput

	game, err := u.store.GetGame(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetGame: %w", err)
	}

	if !game.IsOwner(input.UserID) {
		return output, domain.ErrForbidden
	}

	versions, err := u.store.ListVersions(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.ListVersions: %w", err)
	}

	events, err := u.store.ListGenEvents(ctx, input.GameID, 0)
	if err != nil {
		return output, fmt.Errorf("store.ListGenEvents: %w", err)
	}

	rating, err := u.store.GetRatingAggregate(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetRatingAggregate: %w", err)
	}

	myRating, err := u.store.GetRating(ctx, input.GameID, input.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return output, fmt.Errorf("store.GetRating: %w", err)
	}

	output.Rating = rating
	if err == nil {
		output.MyRating = int(myRating.Value)
	}

	output.Status = string(game.Status)
	output.Stages = buildStages(events)
	output.Spec = game.Spec
	output.Config = game.Config
	output.Versions = make([]dto.Version, 0, len(versions))

	for _, version := range versions {
		output.Versions = append(output.Versions, dto.Version{
			ID:            version.ID,
			Number:        version.Number,
			Source:        string(version.Source),
			Note:          version.Note,
			TokensCost:    version.TokensCost,
			ModelCalls:    version.ModelCalls,
			PromptVersion: version.PromptVersion,
			CreatedAt:     version.CreatedAt,
		})

		output.Cost.Tokens += version.TokensCost
		output.Cost.Calls += version.ModelCalls

		if version.Source == domain.VersionSourceRepair {
			output.Cost.Fixes++
		}
	}

	return output, nil
}

func buildStages(events []domain.GenEvent) []dto.Stage {
	stages := make([]dto.Stage, 0)
	index := make(map[string]int)

	for _, event := range events {
		if event.Type != domain.GenEventStage {
			continue
		}

		var stage dto.Stage

		if err := json.Unmarshal(event.Payload, &stage); err != nil {
			continue
		}

		if pos, ok := index[stage.Stage]; ok {
			stages[pos] = stage
			continue
		}

		index[stage.Stage] = len(stages)
		stages = append(stages, stage)
	}

	return stages
}
