package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestDeleteGameMarksGameDeleted(t *testing.T) {
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
	env.store.EXPECT().DeleteGame(mock.Anything, mock.MatchedBy(func(game domain.Game) bool {
		return game.ID == gameID && game.DeletedAt != nil
	})).Return(nil)

	req := env.authorized(httptest.NewRequest(http.MethodDelete, "/games/"+gameID.String(), nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.DeleteGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if !output.Deleted {
		t.Error("deleted = false, want true")
	}
}

func TestDeleteGameIsForbiddenForForeignGame(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodDelete, "/games/"+gameID.String(), nil))
	rec := env.do(req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestUpdateVisibilityPublishesReadyGame(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:         gameID,
		OwnerID:    userID,
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPrivate,
	}, nil)
	env.store.EXPECT().UpdateGameVisibility(mock.Anything, mock.MatchedBy(func(game domain.Game) bool {
		return game.Visibility == domain.VisibilityPublic
	})).Return(nil)

	body := strings.NewReader(`{"visibility":"public"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPatch, "/games/"+gameID.String()+"/visibility", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.UpdateGameVisibilityOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Visibility != string(domain.VisibilityPublic) {
		t.Errorf("visibility = %q, want public", output.Visibility)
	}
}

func TestUpdateVisibilityRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)

	body := strings.NewReader(`{"visibility":"secret"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPatch, "/games/"+uuid.New().String()+"/visibility", body))
	rec := env.do(req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
