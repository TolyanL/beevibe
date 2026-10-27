package v1

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestStreamGameWritesEventsAndStops(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	events := []domain.GenEvent{
		{Seq: 1, GameID: gameID, Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"running"}`)},
		{Seq: 2, GameID: gameID, Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
	}

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusCoding,
	}, nil).Once()
	env.store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return(events, nil).Once()
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil).Maybe()
	env.store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(2)).Return(nil, nil).Maybe()

	req := env.authorized(httptest.NewRequest(http.MethodGet, env.gamePath(gameID)+"/stream", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("content type = %q, want text/event-stream", contentType)
	}

	body := rec.Body.String()

	for _, want := range []string{"id: 1", "id: 2", "event: stage", `"stage":"analyzing"`} {
		if !strings.Contains(body, want) {
			t.Errorf("stream body misses %q:\n%s", want, body)
		}
	}
}

func TestStreamGameIsForbiddenForForeignGame(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusCoding,
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, env.gamePath(gameID)+"/stream", nil))
	rec := env.do(req)

	body := rec.Body.String()

	if !strings.Contains(body, "event: error") || !strings.Contains(body, "forbidden") {
		t.Fatalf("body = %q, want error event with forbidden", body)
	}
}

func TestStreamGameRejectsUnknownReference(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGameBySlug(mock.Anything, "not-a-uuid").Return(domain.Game{}, domain.ErrNotFound)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games/not-a-uuid/stream", nil))
	rec := env.do(req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
