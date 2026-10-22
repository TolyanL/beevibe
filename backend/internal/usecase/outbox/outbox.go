package outbox

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/pkg/tracer"
)

// Store — порт доступа к таблице outbox.
type Store interface {
	InsertOutboxEvent(ctx context.Context, event domain.Event) error
	ReadOutboxEvents(ctx context.Context, limit int) ([]domain.Event, error)
	MarkOutboxPublished(ctx context.Context, ids []uuid.UUID) error
	InsertGameEvent(ctx context.Context, record domain.GameEventRecord) error
}

// Kafka — порт публикации события в брокер.
type Kafka interface {
	Produce(ctx context.Context, topic, key string, payload []byte) error
}

type UseCase struct {
	store Store
	kafka Kafka
}

func New(store Store, kafka Kafka) *UseCase {
	return &UseCase{store: store, kafka: kafka}
}

// PublishBatch — публикует пачку неотправленных событий и помечает их отправленными.
func (u *UseCase) PublishBatch(ctx context.Context, limit int) (int, error) {
	ctx, span := tracer.Start(ctx, "usecase outbox PublishBatch")
	defer span.End()

	events, err := u.store.ReadOutboxEvents(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("store.ReadOutboxEvents: %w", err)
	}

	if len(events) == 0 {
		return 0, nil
	}

	ids := make([]uuid.UUID, 0, len(events))

	for _, event := range events {
		if err := u.kafka.Produce(ctx, event.Topic, event.Key, event.Payload); err != nil {
			return len(ids), fmt.Errorf("kafka.Produce: %w", err)
		}

		ids = append(ids, event.ID)
	}

	if err := u.store.MarkOutboxPublished(ctx, ids); err != nil {
		return len(ids), fmt.Errorf("store.MarkOutboxPublished: %w", err)
	}

	return len(ids), nil
}
