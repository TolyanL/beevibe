package event

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	eventmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/event"
)

func newTestUseCase(t *testing.T) (*UseCase, *eventmocks.MockStore) {
	t.Helper()

	store := eventmocks.NewMockStore(t)

	return New(store), store
}

func TestHandleEventStoresRecord(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	eventID := uuid.New()
	payload := []byte(`{"id":"` + eventID.String() + `","type":"game.ready"}`)

	store.EXPECT().InsertGameEvent(mock.Anything, mock.MatchedBy(func(record domain.GameEventRecord) bool {
		return record.EventID == eventID &&
			record.Topic == domain.GameEventsTopic &&
			record.EventType == "game.ready" &&
			string(record.Payload) == string(payload)
	})).Return(nil)

	if err := uc.HandleEvent(context.Background(), domain.GameEventsTopic, eventID.String(), payload); err != nil {
		t.Fatalf("HandleEvent() error = %v, want nil", err)
	}
}

func TestHandleEventIgnoresDuplicate(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	eventID := uuid.New()
	payload := []byte(`{"id":"` + eventID.String() + `","type":"game.ready"}`)

	store.EXPECT().InsertGameEvent(mock.Anything, mock.Anything).Return(domain.ErrAlreadyExists)

	if err := uc.HandleEvent(context.Background(), domain.GameEventsTopic, eventID.String(), payload); err != nil {
		t.Fatalf("HandleEvent() error = %v, want nil", err)
	}
}

func TestHandleEventRejectsMalformedPayload(t *testing.T) {
	t.Parallel()

	uc, _ := newTestUseCase(t)

	err := uc.HandleEvent(context.Background(), domain.GameEventsTopic, "key", []byte("not-json"))
	if err == nil {
		t.Fatal("HandleEvent() error = nil, want error")
	}
}

func TestHandleEventRejectsPayloadWithoutID(t *testing.T) {
	t.Parallel()

	uc, _ := newTestUseCase(t)

	err := uc.HandleEvent(context.Background(), domain.GameEventsTopic, "key", []byte(`{"type":"game.ready"}`))
	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("HandleEvent() error = %v, want ErrInvalidInput", err)
	}
}

func TestHandleEventFailsWhenStoreFails(t *testing.T) {
	t.Parallel()

	uc, store := newTestUseCase(t)

	eventID := uuid.New()
	payload := []byte(`{"id":"` + eventID.String() + `","type":"game.ready"}`)

	store.EXPECT().InsertGameEvent(mock.Anything, mock.Anything).Return(errors.New("db is down"))

	err := uc.HandleEvent(context.Background(), domain.GameEventsTopic, eventID.String(), payload)
	if err == nil {
		t.Fatal("HandleEvent() error = nil, want error")
	}
}
