package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
)

func TestCancelGameRejectsIdleGame(t *testing.T) {
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

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/cancel", nil))
	rec := env.do(req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	var envelope errorBody
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if envelope.Error.Code != "nothing_to_cancel" {
		t.Errorf("code = %q, want nothing_to_cancel", envelope.Error.Code)
	}
}
