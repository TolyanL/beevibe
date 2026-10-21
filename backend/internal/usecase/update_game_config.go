package usecase

import (
	"context"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

func (u *UseCase) UpdateGameConfig(ctx context.Context, input dto.UpdateGameConfigInput) (dto.UpdateGameConfigOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase UpdateGameConfig")
	defer span.End()

	var output dto.UpdateGameConfigOutput

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

	latest, err := u.store.GetLatestVersion(ctx, input.GameID)
	if err != nil {
		return output, fmt.Errorf("store.GetLatestVersion: %w", err)
	}

	current := domain.GameConfig{}

	if game.Config != nil {
		current = *game.Config
	}

	update := domain.GameConfigUpdate{}

	if input.Tempo != nil {
		tempo := domain.Tempo(*input.Tempo)
		update.Tempo = &tempo
	}

	if input.Palette != nil {
		palette := input.Palette
		update.Palette = &palette
	}

	if input.MapSize != nil {
		mapSize := domain.MapSize(*input.MapSize)
		update.MapSize = &mapSize
	}

	if input.Difficulty != nil {
		difficulty := *input.Difficulty
		update.Difficulty = &difficulty
	}

	cfg, err := current.Apply(update)
	if err != nil {
		return output, fmt.Errorf("config.Apply: %w", err)
	}

	html, err := u.validator.ApplyConfig(ctx, latest.HTML, cfg)
	if err != nil {
		return output, fmt.Errorf("validator.ApplyConfig: %w", err)
	}

	number, err := u.nextVersionNumber(ctx, input.GameID)
	if err != nil {
		return output, err
	}

	version, err := domain.NewVersion(input.GameID, number, html, domain.VersionSourceConfig)
	if err != nil {
		return output, fmt.Errorf("domain.NewVersion: %w", err)
	}

	version.Note = "config update"

	if err := u.store.CreateVersion(ctx, version); err != nil {
		return output, fmt.Errorf("store.CreateVersion: %w", err)
	}

	if err := game.ApplyConfig(cfg); err != nil {
		return output, fmt.Errorf("game.ApplyConfig: %w", err)
	}

	if err := u.store.UpdateGame(ctx, game); err != nil {
		return output, fmt.Errorf("store.UpdateGame: %w", err)
	}

	if err := u.emitArtifactEvent(ctx, input.GameID, "config", cfg); err != nil {
		return output, fmt.Errorf("u.emitArtifactEvent: %w", err)
	}

	output.VersionID = version.ID

	return output, nil
}
