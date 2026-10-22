package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

// Store — порт записи в проекцию game_event_log.
type Store interface {
	InsertGameEvent(ctx context.Context, record domain.GameEventRecord) error
}

type UseCase struct {
	store Store
}

func New(store Store) *UseCase {
	return &UseCase{store: store}
}

// Handle — разбор и идемпотентная запись события из Kafka: повторная доставка
// того же event_id не создаёт дубль (уникальный индекс по event_id).
func (u *UseCase) HandleEvent(ctx context.Context, topic, key string, payload []byte) error {
	ctx, span := tracer.Start(ctx, "usecase event HandleEvent")
	defer span.End()

	var body domain.GameEventPayload

	if err := json.Unmarshal(payload, &body); err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	if body.ID == uuid.Nil {
		return fmt.Errorf("%w: event payload has no id", domain.ErrInvalidInput)
	}

	record := domain.GameEventRecord{
		ID:         uuid.New(),
		EventID:    body.ID,
		Topic:      topic,
		Key:        key,
		EventType:  body.Type,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}

	if err := u.store.InsertGameEvent(ctx, record); err != nil {
		if errors.Is(err, domain.ErrAlreadyExists) {
			return nil
		}

		return fmt.Errorf("store.InsertGameEvent: %w", err)
	}

	return nil
}
