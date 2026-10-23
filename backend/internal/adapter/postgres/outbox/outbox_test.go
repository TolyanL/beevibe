package outbox

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
	"github.com/TolyanL/beevibe/internal/domain"
)

func TestReadOutboxEventsScansRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(10)).Return(pgtest.Rows(t,
		[]any{uuid.New(), "game.events", "game-1", []byte(`{"a":1}`), createdAt},
		[]any{uuid.New(), "game.events", "game-2", []byte(`{"a":2}`), createdAt},
	), nil)

	events, err := postgres.ReadOutboxEvents(ctx, 10)
	if err != nil {
		t.Fatalf("ReadOutboxEvents() error = %v, want nil", err)
	}

	if len(events) != 2 {
		t.Fatalf("events = %d, want 2", len(events))
	}

	if events[0].Topic != "game.events" || events[0].Key != "game-1" || string(events[0].Payload) != `{"a":1}` {
		t.Errorf("first event = %+v, want game.events game-1", events[0])
	}
}

func TestReadOutboxEventsReturnsEmptySlice(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Rows(t), nil)

	events, err := postgres.ReadOutboxEvents(ctx, 10)
	if err != nil {
		t.Fatalf("ReadOutboxEvents() error = %v, want nil", err)
	}

	if events == nil || len(events) != 0 {
		t.Fatalf("events = %v, want empty slice", events)
	}
}

func TestMarkOutboxPublishedUpdatesRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	ids := []uuid.UUID{uuid.New(), uuid.New()}

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(ids)).Return(pgtest.Tag(2), nil)

	if err := postgres.MarkOutboxPublished(ctx, ids); err != nil {
		t.Fatalf("MarkOutboxPublished() error = %v, want nil", err)
	}
}

func TestInsertGameEventStoresRecord(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	record := domain.GameEventRecord{
		ID:         uuid.New(),
		EventID:    uuid.New(),
		Topic:      "game.events",
		Key:        "game-1",
		EventType:  "game.ready",
		Payload:    []byte(`{"id":"event-1"}`),
		ReceivedAt: time.Now().UTC(),
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(
		record.ID, record.EventID, record.Topic, record.Key, string(record.EventType),
		string(record.Payload), mock.Anything,
	)).Return(pgtest.Tag(1), nil)

	if err := postgres.InsertGameEvent(ctx, record); err != nil {
		t.Fatalf("InsertGameEvent() error = %v, want nil", err)
	}
}

func TestInsertGameEventMapsDuplicateToAlreadyExists(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.Tag(0), pgtest.UniqueViolation("game_event_log_event_id_key"))

	err := postgres.InsertGameEvent(ctx, domain.GameEventRecord{})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("InsertGameEvent() error = %v, want ErrAlreadyExists", err)
	}
}

func TestInsertGameEventMapsSkippedConflictToAlreadyExists(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Tag(0), nil)

	err := postgres.InsertGameEvent(ctx, domain.GameEventRecord{})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("InsertGameEvent() error = %v, want ErrAlreadyExists", err)
	}
}

func TestInsertOutboxEventStoresEvent(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	event := domain.Event{
		ID:        uuid.New(),
		Topic:     "game.events",
		Key:       "game-1",
		Payload:   []byte(`{"a":1}`),
		CreatedAt: time.Now().UTC(),
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(
		event.ID, event.Topic, event.Key, string(event.Payload), mock.Anything,
	)).Return(pgtest.Tag(1), nil)

	if err := postgres.InsertOutboxEvent(ctx, event); err != nil {
		t.Fatalf("InsertOutboxEvent() error = %v, want nil", err)
	}
}

func TestInsertOutboxEventWrapsDriverError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Exec(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.Tag(0), errors.New("connection reset"))

	if err := postgres.InsertOutboxEvent(ctx, domain.Event{}); err == nil {
		t.Fatal("InsertOutboxEvent() error = nil, want error")
	}
}
