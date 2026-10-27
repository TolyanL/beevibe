package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

const (
	metaStage          = "meta"
	stageStatusRunning = "running"
	stageStatusDone    = "done"
)

type artifact struct {
	Kind    string `json:"kind"`
	Content any    `json:"content"`
}

func (u *UseCase) emitStageEvent(ctx context.Context, gameID uuid.UUID, stage domain.Status, status string, durationMs int64, issues int) error {
	return u.emitStage(ctx, gameID, dto.Stage{
		Stage:      string(stage),
		Status:     status,
		DurationMs: durationMs,
		Issues:     issues,
	})
}

func (u *UseCase) emitOperationStageEvent(ctx context.Context, gameID uuid.UUID, kind domain.JobKind, status string) error {
	return u.emitStage(ctx, gameID, dto.Stage{
		Stage:     string(domain.StatusPatching),
		Status:    status,
		Operation: kind.String(),
	})
}

func (u *UseCase) emitStage(ctx context.Context, gameID uuid.UUID, payload dto.Stage) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	event := domain.GenEvent{
		GameID:    gameID,
		Type:      domain.GenEventStage,
		Payload:   raw,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.store.InsertGenEvent(ctx, event); err != nil {
		return fmt.Errorf("store.InsertGenEvent: %w", err)
	}

	return nil
}

func (u *UseCase) emitArtifactEvent(ctx context.Context, gameID uuid.UUID, kind string, content any) error {
	raw, err := json.Marshal(artifact{Kind: kind, Content: content})
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	event := domain.GenEvent{
		GameID:    gameID,
		Type:      domain.GenEventArtifact,
		Payload:   raw,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.store.InsertGenEvent(ctx, event); err != nil {
		return fmt.Errorf("store.InsertGenEvent: %w", err)
	}

	return nil
}

func (u *UseCase) emitDoneEvent(ctx context.Context, gameID uuid.UUID, versionID uuid.UUID, tokens, calls int) error {
	payload := map[string]any{
		"versionId": versionID.String(),
		"cost": map[string]int{
			"tokens": tokens,
			"calls":  calls,
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	event := domain.GenEvent{
		GameID:    gameID,
		Type:      domain.GenEventDone,
		Payload:   raw,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.store.InsertGenEvent(ctx, event); err != nil {
		return fmt.Errorf("store.InsertGenEvent: %w", err)
	}

	return nil
}

func (u *UseCase) emitErrorEvent(ctx context.Context, gameID uuid.UUID, code, message string) error {
	payload := map[string]string{"code": code, "message": message}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	event := domain.GenEvent{
		GameID:    gameID,
		Type:      domain.GenEventError,
		Payload:   raw,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.store.InsertGenEvent(ctx, event); err != nil {
		return fmt.Errorf("store.InsertGenEvent: %w", err)
	}

	return nil
}
