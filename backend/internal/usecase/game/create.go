package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
	"github.com/TolyanL/beevibe/pkg/tracer"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const operationCreateGame = "create_game"

func (u *UseCase) Create(ctx context.Context, input dto.CreateGameInput) (dto.CreateGameOutput, error) {
	ctx, span := tracer.Start(ctx, "usecase game Create")
	defer span.End()

	var output dto.CreateGameOutput

	if err := input.Validate(); err != nil {
		return output, fmt.Errorf("input.Validate: %w", err)
	}

	if input.IdempotencyKey != "" {
		replay, err := u.replayCreateGame(ctx, input)
		if err != nil {
			return output, err
		}

		if replay != nil {
			return *replay, nil
		}
	}

	game, err := domain.NewGame(input.UserID, input.Prompt, buildParams(input))
	if err != nil {
		return output, fmt.Errorf("domain.NewGame: %w", err)
	}

	err = transaction.Wrap(ctx, func(ctx context.Context) error {
		user, err := u.store.GetUserByIDForUpdate(ctx, input.UserID)
		if err != nil {
			return fmt.Errorf("store.GetUserByIDForUpdate: %w", err)
		}

		if err := user.Debit(u.tokens.PriceGenerate); err != nil {
			return fmt.Errorf("user.Debit: %w", err)
		}

		if err := u.store.UpdateUser(ctx, user); err != nil {
			return fmt.Errorf("store.UpdateUser: %w", err)
		}

		if err := u.store.CreateGame(ctx, game); err != nil {
			return fmt.Errorf("store.CreateGame: %w", err)
		}

		op, err := domain.NewTokenOp(user.ID, &game.ID, -u.tokens.PriceGenerate, domain.ReasonGenerate)
		if err != nil {
			return fmt.Errorf("domain.NewTokenOp: %w", err)
		}

		if err := u.store.InsertTokenOp(ctx, op); err != nil {
			return fmt.Errorf("store.InsertTokenOp: %w", err)
		}

		if err := u.publishEvent(ctx, game, domain.EventGameCreated, nil, 0, ""); err != nil {
			return fmt.Errorf("u.publishEvent: %w", err)
		}

		return nil
	})
	if err != nil {
		u.releaseIdempotencyKey(ctx, input)

		return output, fmt.Errorf("transaction.Wrap: %w", err)
	}

	if input.IdempotencyKey != "" {
		if err := u.store.CompleteIdempotencyKey(ctx, input.IdempotencyKey, operationCreateGame, game.ID.String()); err != nil {
			return output, fmt.Errorf("store.CompleteIdempotencyKey: %w", err)
		}
	}

	return dto.CreateGameOutput{ID: game.ID, Status: string(game.Status)}, nil
}

func (u *UseCase) replayCreateGame(ctx context.Context, input dto.CreateGameInput) (*dto.CreateGameOutput, error) {
	inserted, err := u.store.InsertIdempotencyKey(ctx, input.IdempotencyKey, operationCreateGame, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("store.InsertIdempotencyKey: %w", err)
	}

	if inserted {
		return nil, nil
	}

	resultID, found, err := u.store.GetIdempotencyResult(ctx, input.IdempotencyKey, operationCreateGame)
	if err != nil {
		return nil, fmt.Errorf("store.GetIdempotencyResult: %w", err)
	}

	if !found || resultID == "" {
		return nil, domain.ErrIdempotencyInProgress
	}

	gameID, err := uuid.Parse(resultID)
	if err != nil {
		return nil, fmt.Errorf("uuid.Parse: %w", err)
	}

	game, err := u.store.GetGame(ctx, gameID)
	if err != nil {
		return nil, fmt.Errorf("store.GetGame: %w", err)
	}

	return &dto.CreateGameOutput{ID: game.ID, Status: string(game.Status)}, nil
}

func (u *UseCase) releaseIdempotencyKey(ctx context.Context, input dto.CreateGameInput) {
	if input.IdempotencyKey == "" {
		return
	}

	if err := u.store.DeleteIdempotencyKey(ctx, input.IdempotencyKey, operationCreateGame); err != nil {
		return
	}
}

func buildParams(input dto.CreateGameInput) domain.GameParams {
	params := domain.GameParams{
		Genre:   domain.Genre(input.Genre),
		MapSize: domain.MapSize(input.MapSize),
		Style:   domain.Style(input.Style),
	}

	if input.Tempo != nil {
		params.Tempo = domain.Tempo(*input.Tempo)
	}

	return params
}
