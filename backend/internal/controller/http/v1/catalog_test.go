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

func TestCatalogReturnsCards(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	gameID := uuid.New()

	env.store.EXPECT().ListCatalog(mock.Anything, dto.CatalogFilter{
		Sort:   "fresh",
		Limit:  20,
		Offset: 0,
	}).Return([]dto.CatalogCard{
		{ID: gameID, Prompt: "змейка", Genre: "snake", Rating: dto.RatingAggregate{Average: 5}},
	}, nil)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/catalog?sort=fresh", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.ListCatalogOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if len(output.Items) != 1 || output.Items[0].ID != gameID {
		t.Errorf("items = %+v, want one card of game %s", output.Items, gameID)
	}
}

func TestCatalogRejectsUnknownSort(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/catalog?sort=unknown", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestProfileReturnsBalanceGamesAndOps(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetUserByID(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		Email:        domain.Email(testEmail),
		TokenBalance: 680_000,
	}, nil)
	env.store.EXPECT().ListUserGamesWithCost(mock.Anything, userID, 50, 0).Return([]dto.ProfileGame{
		{ID: gameID, Prompt: "змейка", Status: "ready", TokensCost: 320_000},
	}, nil)
	env.store.EXPECT().ListTokenOps(mock.Anything, userID, 50).Return([]domain.TokenOp{
		{UserID: userID, GameID: &gameID, Delta: -320_000, Reason: domain.ReasonGenerate},
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/profile", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.GetProfileOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Balance != 680_000 {
		t.Errorf("balance = %d, want 680000", output.Balance)
	}

	if len(output.Games) != 1 || len(output.TokenOps) != 1 {
		t.Errorf("games = %d ops = %d, want one of each", len(output.Games), len(output.TokenOps))
	}

	if output.Prices.Generate != 320_000 || output.Prices.Patch != 40_000 {
		t.Errorf("prices = %+v, want generate 320000 patch 40000", output.Prices)
	}
}

func TestTopupCreditsBalance(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 100,
	}, nil)
	env.store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 100+500_000
	})).Return(nil)
	env.store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 500_000 && op.Reason == domain.ReasonTopup
	})).Return(nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/profile/topup", strings.NewReader(`{}`)))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.TopupTokensOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Balance != 100+500_000 {
		t.Errorf("balance = %d, want %d", output.Balance, int64(100+500_000))
	}
}

func TestHealthReturnsStatus(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/health", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.HealthOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Status != "ok" {
		t.Errorf("status = %q, want ok", output.Status)
	}
}
