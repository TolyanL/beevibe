package outbox

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	outboxmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/outbox"
)

func newTestUseCase(t *testing.T) (*UseCase, *outboxmocks.MockStore, *outboxmocks.MockKafka) {
	t.Helper()

	store := outboxmocks.NewMockStore(t)
	kafka := outboxmocks.NewMockKafka(t)

	return New(store, kafka), store, kafka
}

func TestPublishBatchProducesEventsAndMarksThem(t *testing.T) {
	t.Parallel()

	uc, store, kafka := newTestUseCase(t)

	events := []domain.Event{
		{ID: uuid.New(), Topic: domain.GameEventsTopic, Key: "game-1", Payload: []byte(`{"a":1}`)},
		{ID: uuid.New(), Topic: domain.GameEventsTopic, Key: "game-2", Payload: []byte(`{"a":2}`)},
	}

	store.EXPECT().ReadOutboxEvents(mock.Anything, 10).Return(events, nil)
	kafka.EXPECT().Produce(mock.Anything, domain.GameEventsTopic, "game-1", events[0].Payload).Return(nil)
	kafka.EXPECT().Produce(mock.Anything, domain.GameEventsTopic, "game-2", events[1].Payload).Return(nil)
	store.EXPECT().MarkOutboxPublished(mock.Anything, []uuid.UUID{events[0].ID, events[1].ID}).Return(nil)

	count, err := uc.PublishBatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("PublishBatch() error = %v, want nil", err)
	}

	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestPublishBatchDoesNothingWhenEmpty(t *testing.T) {
	t.Parallel()

	uc, store, _ := newTestUseCase(t)

	store.EXPECT().ReadOutboxEvents(mock.Anything, 10).Return([]domain.Event{}, nil)

	count, err := uc.PublishBatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("PublishBatch() error = %v, want nil", err)
	}

	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

func TestPublishBatchStopsOnKafkaError(t *testing.T) {
	t.Parallel()

	uc, store, kafka := newTestUseCase(t)

	events := []domain.Event{
		{ID: uuid.New(), Topic: domain.GameEventsTopic, Key: "game-1", Payload: []byte(`{"a":1}`)},
		{ID: uuid.New(), Topic: domain.GameEventsTopic, Key: "game-2", Payload: []byte(`{"a":2}`)},
	}

	store.EXPECT().ReadOutboxEvents(mock.Anything, 10).Return(events, nil)
	kafka.EXPECT().Produce(mock.Anything, domain.GameEventsTopic, "game-1", events[0].Payload).Return(nil)
	kafka.EXPECT().Produce(mock.Anything, domain.GameEventsTopic, "game-2", events[1].Payload).Return(errors.New("kafka is down"))

	_, err := uc.PublishBatch(context.Background(), 10)
	if err == nil {
		t.Fatal("PublishBatch() error = nil, want error")
	}
}
