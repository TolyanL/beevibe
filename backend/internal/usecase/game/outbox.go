package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
)

func (u *UseCase) publishEvent(
	ctx context.Context,
	game domain.Game,
	eventType domain.EventType,
	versionID *uuid.UUID,
	tokensCost int,
	failure string,
) error {
	payload := domain.GameEventPayload{
		ID:         uuid.New(),
		Type:       eventType,
		GameID:     game.ID,
		OwnerID:    game.OwnerID,
		Prompt:     game.Prompt,
		Status:     game.Status,
		VersionID:  versionID,
		TokensCost: tokensCost,
		Error:      failure,
		OccurredAt: time.Now().UTC(),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	event := domain.Event{
		ID:        payload.ID,
		Type:      eventType,
		Topic:     domain.GameEventsTopic,
		Key:       game.ID.String(),
		Payload:   raw,
		CreatedAt: payload.OccurredAt,
	}

	if err := u.outbox.InsertOutboxEvent(ctx, event); err != nil {
		return fmt.Errorf("outbox.InsertOutboxEvent: %w", err)
	}

	return nil
}
