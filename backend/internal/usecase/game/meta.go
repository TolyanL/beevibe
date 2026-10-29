package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const operationGenerateMeta = "generate_meta"

func (u *UseCase) UpdateMeta(ctx context.Context, input dto.UpdateGameMetaInput) (dto.UpdateGameMetaOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game UpdateMeta")
	defer span.End()

	var output dto.UpdateGameMetaOutput

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

	title := game.Title
	if input.Title != nil {
		title = *input.Title
	}

	description := game.Description
	if input.Description != nil {
		description = *input.Description
	}

	cover := game.Cover
	if input.Cover != nil {
		cover = input.Cover
	}

	if err := game.SetMeta(title, description, cover); err != nil {
		return output, fmt.Errorf("game.SetMeta: %w", err)
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.UpdateGameMeta(ctx, game); err != nil {
			return fmt.Errorf("store.UpdateGameMeta: %w", err)
		}

		return nil
	}); err != nil {
		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	output.Title = game.Title
	output.Description = game.Description
	output.Cover = game.Cover

	return output, nil
}

func (u *UseCase) GenerateMeta(ctx context.Context, input dto.GenerateGameMetaInput) (dto.GenerateGameMetaOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game GenerateMeta")
	defer span.End()

	var output dto.GenerateGameMetaOutput

	if input.IdempotencyKey != "" {
		_, replayed, err := u.claimIdempotencyKey(ctx, input.IdempotencyKey, operationGenerateMeta, input.UserID)
		if err != nil {
			return output, err
		}

		if replayed {
			output.Status = string(domain.StatusPatching)

			return output, nil
		}
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

	job, err := domain.NewGameJob(game.ID, game.OwnerID, domain.JobKindMeta, "")
	if err != nil {
		return output, fmt.Errorf("domain.NewGameJob: %w", err)
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		user, err := u.store.GetUserByIDForUpdate(ctx, input.UserID)
		if err != nil {
			return fmt.Errorf("store.GetUserByIDForUpdate: %w", err)
		}

		if err := user.Debit(u.tokens.PriceMeta); err != nil {
			return fmt.Errorf("user.Debit: %w", err)
		}

		if err := u.store.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("store.UpdateUser: %w", err)
		}

		op, err := domain.NewTokenOp(user.ID, &game.ID, -u.tokens.PriceMeta, domain.ReasonMeta)
		if err != nil {
			return fmt.Errorf("domain.NewTokenOp: %w", err)
		}

		if err := u.store.InsertTokenOp(ctx, op); err != nil {
			return fmt.Errorf("store.InsertTokenOp: %w", err)
		}

		if err := u.store.InsertGameJob(ctx, job); err != nil {
			return fmt.Errorf("store.InsertGameJob: %w", err)
		}

		return u.emitStage(ctx, game.ID, dto.Stage{
			Stage:     metaStage,
			Status:    stageStatusRunning,
			Operation: string(domain.JobKindMeta),
		})
	})
	if err != nil {
		u.releaseIdempotencyKey(ctx, input.IdempotencyKey, operationGenerateMeta)

		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if err := u.completeIdempotencyKey(ctx, input.IdempotencyKey, operationGenerateMeta, game.ID.String()); err != nil {
		return output, err
	}

	output.Status = "queued"

	return output, nil
}

func (u *UseCase) runMetaJob(ctx context.Context, job domain.GameJob, game domain.Game) error {
	meta, tokens, calls, err := u.requestMeta(ctx, game)
	if err != nil {
		return u.abortMetaJob(ctx, game, err)
	}

	if err := game.SetMeta(meta.Title, meta.Description, meta.Cover); err != nil {
		return u.abortMetaJob(ctx, game, fmt.Errorf("game.SetMeta: %w", err))
	}

	if err := transaction.Wrap(ctx, func(ctx context.Context) error {
		if err := u.store.UpdateGameMeta(ctx, game); err != nil {
			return fmt.Errorf("store.UpdateGameMeta: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("transaction.Wrap: %w", err)
	}

	if err := u.emitStage(ctx, game.ID, dto.Stage{
		Stage:     metaStage,
		Status:    "done",
		Operation: string(domain.JobKindMeta),
	}); err != nil {
		return fmt.Errorf("u.emitStage: %w", err)
	}

	payload := map[string]any{
		"title":       game.Title,
		"description": game.Description,
		"cover":       game.Cover,
	}

	if err := u.emitArtifactEvent(ctx, game.ID, "meta", payload); err != nil {
		return fmt.Errorf("u.emitArtifactEvent: %w", err)
	}

	if err := u.emitDoneEvent(ctx, game.ID, job.ID, tokens, calls); err != nil {
		return fmt.Errorf("u.emitDoneEvent: %w", err)
	}

	return nil
}

func (u *UseCase) abortMetaJob(ctx context.Context, game domain.Game, cause error) error {
	failCtx := context.WithoutCancel(ctx)

	if !u.replied(game.ID) {
		if err := transaction.Wrap(failCtx, func(ctx context.Context) error {
			return u.refundInTx(ctx, game.OwnerID, game.ID, u.tokens.PriceMeta)
		}); err != nil {
			return errors.Join(cause, fmt.Errorf("transaction.Wrap: %w", err))
		}
	}

	if err := u.emitStage(failCtx, game.ID, dto.Stage{
		Stage:     metaStage,
		Status:    "error",
		Operation: string(domain.JobKindMeta),
	}); err != nil {
		return errors.Join(cause, fmt.Errorf("u.emitStage: %w", err))
	}

	if err := u.emitErrorEvent(failCtx, game.ID, errorCode(cause), cause.Error()); err != nil {
		return errors.Join(cause, fmt.Errorf("u.emitErrorEvent: %w", err))
	}

	return cause
}

func (u *UseCase) requestMeta(ctx context.Context, game domain.Game) (domain.GameMeta, int, int, error) {
	system, err := u.llm.Prompt("meta")
	if err != nil {
		return domain.GameMeta{}, 0, 0, fmt.Errorf("llm.Prompt: %w", err)
	}

	p := &pipeline{game: game}

	raw, err := u.complete(ctx, p, system, buildMetaPrompt(game), temperatureIdeas, u.reasoningMeta())
	if err != nil {
		return domain.GameMeta{}, p.tokens, p.calls, fmt.Errorf("u.complete: %w", err)
	}

	var payload struct {
		Title       string            `json:"title"`
		Description string            `json:"description"`
		Cover       *domain.GameCover `json:"cover"`
	}

	object, err := extractJSONObject(raw)
	if err != nil {
		return domain.GameMeta{}, p.tokens, p.calls, fmt.Errorf("extractJSONObject: %w", err)
	}

	if err := json.Unmarshal([]byte(object), &payload); err != nil {
		return domain.GameMeta{}, p.tokens, p.calls, fmt.Errorf("json.Unmarshal: %w", err)
	}

	meta, err := domain.NewGameMeta(payload.Title, payload.Description, payload.Cover)
	if err != nil {
		return domain.GameMeta{}, p.tokens, p.calls, fmt.Errorf("domain.NewGameMeta: %w", err)
	}

	return meta, p.tokens, p.calls, nil
}

func coverOf(game domain.Game) *domain.GameCover {
	if game.Cover != nil {
		return game.Cover
	}

	if game.Spec == nil {
		return nil
	}

	return domain.CoverFromPalette(game.Spec.Palette, string(game.Params.Style), game.DisplaySlug())
}

func (u *UseCase) applyCoverFromSpec(ctx context.Context, game *domain.Game) error {
	if game.Cover != nil || game.Spec == nil {
		return nil
	}

	cover := domain.CoverFromPalette(game.Spec.Palette, string(game.Params.Style), game.DisplaySlug())
	if cover == nil {
		return nil
	}

	game.Cover = cover

	return nil
}
