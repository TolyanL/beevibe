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

func TestAcceptReviewQueuesAssembly(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusNeedsDecision,
	}, nil)
	env.store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusQueued, "").Return(nil)
	env.store.EXPECT().SaveGameResumeStage(mock.Anything, gameID, domain.StatusAssembling).Return(nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/review/accept", nil))
	rec := env.do(req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	var output dto.ResolveReviewOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.ResumeFrom != string(domain.StatusAssembling) {
		t.Errorf("resumeFrom = %q, want assembling", output.ResumeFrom)
	}
}

func TestRegenerateReviewQueuesCoding(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusNeedsDecision,
	}, nil)
	env.store.EXPECT().UpdateGameStatus(mock.Anything, gameID, domain.StatusQueued, "").Return(nil)
	env.store.EXPECT().SaveGameResumeStage(mock.Anything, gameID, domain.StatusCoding).Return(nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/review/regenerate", nil))
	rec := env.do(req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusAccepted, rec.Body.String())
	}

	var output dto.ResolveReviewOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.ResumeFrom != string(domain.StatusCoding) {
		t.Errorf("resumeFrom = %q, want coding", output.ResumeFrom)
	}
}
