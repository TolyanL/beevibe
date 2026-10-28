package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListUserGamesReturnsPages(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().ListUserGames(mock.Anything, userID, dto.ListUserGamesInput{
		Status: "ready",
		Sort:   "tokens_cost",
		Limit:  5,
		Offset: 10,
	}).Return([]dto.ProfileGame{
		{ID: gameID, Prompt: "змейка", Status: "ready", TokensCost: 320},
	}, 1, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games?status=ready&sort=tokens_cost&limit=5&offset=10", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.ListUserGamesOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if len(output.Items) != 1 || output.Items[0].ID != gameID || output.Total != 1 {
		t.Errorf("output = %+v, want one game with total 1", output)
	}
}

func TestListUserGamesRejectsUnknownStatus(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games?status=released", nil))
	rec := env.do(req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestListUserGamesRequiresAuth(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/games", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
