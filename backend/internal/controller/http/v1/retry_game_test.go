package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestRetryGameRequeuesFailedGame(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    userID,
		Status:     domain.StatusFailed,
		FailReason: "llm unavailable",
	}, nil)
	env.store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusQueued, "").Return(nil)
	env.store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return([]domain.GenEvent{
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"planning","status":"done"}`)},
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"coding","status":"error"}`)},
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/retry", nil))
	rec := env.do(req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	var output dto.RetryGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Status != string(domain.StatusQueued) {
		t.Errorf("status = %q, want queued", output.Status)
	}

	if output.ResumeFrom != string(domain.StatusCoding) {
		t.Errorf("resumeFrom = %q, want coding (the stage that failed)", output.ResumeFrom)
	}
}

func TestRetryGameRejectsReadyGame(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/retry", nil))
	rec := env.do(req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
