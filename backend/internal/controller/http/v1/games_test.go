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
	"github.com/TolyanL/beevibe/pkg/llm"
)

func TestCreateGameDebitsTokens(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 1_000_000,
	}, nil)
	env.store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 1_000_000-320_000
	})).Return(nil)
	env.store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -320_000 && op.Reason == domain.ReasonGenerate
	})).Return(nil)
	env.store.EXPECT().CreateGame(mock.Anything, mock.MatchedBy(func(game domain.Game) bool {
		return game.OwnerID == userID && game.Status == domain.StatusQueued
	})).Return(nil)
	env.store.EXPECT().InsertOutboxEvent(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
		return event.Type == domain.EventGameCreated
	})).Return(nil)

	body := strings.NewReader(`{"prompt":"змейка","genre":"snake","tempo":6}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games", body))
	rec := env.do(req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var output dto.CreateGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.ID == uuid.Nil {
		t.Error("id = nil, want uuid")
	}

	if output.Status != string(domain.StatusQueued) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusQueued)
	}
}

func TestCreateGameRejectsNotEnoughTokens(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 10,
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(`{"prompt":"змейка"}`)))
	rec := env.do(req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestCreateGameReplaysIdempotencyKey(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().InsertIdempotencyKey(mock.Anything, "key-1", "create_game", userID).Return(false, nil)
	env.store.EXPECT().GetIdempotencyResult(mock.Anything, "key-1", "create_game").Return(gameID.String(), true, nil)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games", strings.NewReader(`{"prompt":"змейка"}`)))
	req.Header.Set("Idempotency-Key", "key-1")
	rec := env.do(req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var output dto.CreateGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.ID != gameID {
		t.Errorf("id = %s, want %s", output.ID, gameID)
	}
}

func TestGetGameReturnsVersionsAndCost(t *testing.T) {
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
	env.store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{
		{ID: uuid.New(), Number: 1, Source: domain.VersionSourceGenerate, TokensCost: 100, ModelCalls: 6},
	}, nil)
	env.store.EXPECT().ListGenEvents(mock.Anything, gameID, int64(0)).Return([]domain.GenEvent{
		{Type: domain.GenEventStage, Payload: []byte(`{"stage":"analyzing","status":"done"}`)},
	}, nil)
	env.store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{Total: 0}, nil)
	env.store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{}, domain.ErrNotFound)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games/"+gameID.String(), nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.GetGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Status != string(domain.StatusReady) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusReady)
	}

	if len(output.Versions) != 1 || output.Cost.Tokens != 100 {
		t.Errorf("versions = %+v cost = %+v, want one version with 100 tokens", output.Versions, output.Cost)
	}
}

func TestGetGameReturnsNotFound(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{}, domain.ErrNotFound)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games/"+gameID.String(), nil))
	rec := env.do(req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetForeignGameIsForbidden(t *testing.T) {
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

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games/"+gameID.String(), nil))
	rec := env.do(req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPatchCreatesVersionAndCharges(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)
	env.store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{latest}, nil)
	env.store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 100_000,
	}, nil)
	env.store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 60_000
	})).Return(nil)
	env.store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -40_000 && op.Reason == domain.ReasonPatch
	})).Return(nil)

	env.llm.EXPECT().Prompt("patch").Return("system prompt", nil)
	env.llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{
		Text:  gameHTML,
		Usage: llm.Usage{PromptTokens: 10, CompletionTokens: 20},
	}, nil)
	env.llm.EXPECT().PromptVersion().Return("v1")
	env.validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil)

	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)
	env.store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourcePatch && version.Number == 2
	})).Return(nil)
	env.store.EXPECT().InsertOutboxEvent(mock.Anything, mock.MatchedBy(func(event domain.Event) bool {
		return event.Type == domain.EventGamePatched
	})).Return(nil)

	body := strings.NewReader(`{"text":"сделай кнопки больше"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/patch", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.PatchGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.VersionID == uuid.Nil || output.TokensSpent != 30 {
		t.Errorf("output = %+v, want version id and 30 tokens", output)
	}
}

func TestPatchIsForbiddenForForeignGame(t *testing.T) {
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

	body := strings.NewReader(`{"text":"добавь уровень"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/patch", body))
	rec := env.do(req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestPatchRejectsGameInProgress(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusCoding,
	}, nil)

	body := strings.NewReader(`{"text":"добавь уровень"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/patch", body))
	rec := env.do(req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestPatchRefundsOnRejectedResult(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil)
	env.store.EXPECT().ListVersions(mock.Anything, gameID).Return([]domain.Version{latest}, nil)
	env.store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 100_000,
	}, nil).Once()
	env.store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 60_000
	})).Return(nil).Once()
	env.store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == -40_000
	})).Return(nil).Once()

	env.llm.EXPECT().Prompt("patch").Return("system prompt", nil)
	env.llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil)
	env.validator.EXPECT().Check(mock.Anything, gameHTML).Return([]domain.Issue{
		{Code: "external_url", Message: "external url found"},
	}, nil)

	env.store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 60_000,
	}, nil).Once()
	env.store.EXPECT().UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
		return user.TokenBalance == 100_000
	})).Return(nil).Once()
	env.store.EXPECT().InsertTokenOp(mock.Anything, mock.MatchedBy(func(op domain.TokenOp) bool {
		return op.Delta == 40_000 && op.Reason == domain.ReasonRefund
	})).Return(nil).Once()

	body := strings.NewReader(`{"text":"сделай кнопки больше"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/patch", body))
	rec := env.do(req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestUpdateConfigCreatesVersion(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
		Config:  &domain.GameConfig{Tempo: 5, MapSize: domain.MapSizeMedium, Style: domain.StyleNeon},
	}, nil)
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil).Once()
	env.validator.EXPECT().ApplyConfig(mock.Anything, gameHTML, mock.MatchedBy(func(cfg domain.GameConfig) bool {
		return cfg.Tempo == 9
	})).Return(gameHTML, nil)
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil).Once()
	env.store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceConfig
	})).Return(nil)
	env.store.EXPECT().UpdateGame(mock.Anything, mock.Anything).Return(nil)
	env.store.EXPECT().InsertGenEvent(mock.Anything, mock.Anything).Return(nil)

	body := strings.NewReader(`{"tempo":9}`)
	req := env.authorized(httptest.NewRequest(http.MethodPatch, "/games/"+gameID.String()+"/config", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestUpdateConfigRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)

	body := strings.NewReader(`{"tempo":99}`)
	req := env.authorized(httptest.NewRequest(http.MethodPatch, "/games/"+gameID.String()+"/config", body))
	rec := env.do(req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRollbackCreatesVersion(t *testing.T) {
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
	env.store.EXPECT().GetVersion(mock.Anything, gameID, 1).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 1,
		HTML:   gameHTML,
	}, nil)
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 3,
		HTML:   gameHTML,
	}, nil)
	env.store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceRollback && version.Number == 4
	})).Return(nil)

	body := strings.NewReader(`{"version":1}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/rollback", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestFixRepairsGame(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	latest := domain.Version{ID: uuid.New(), GameID: gameID, Number: 1, HTML: gameHTML}

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: userID,
		Status:  domain.StatusReady,
	}, nil)
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil).Once()
	env.validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil).Once()
	env.llm.EXPECT().Prompt("repair").Return("system prompt", nil)
	env.llm.EXPECT().Complete(mock.Anything, mock.Anything).Return(llm.Response{Text: gameHTML}, nil)
	env.validator.EXPECT().Check(mock.Anything, gameHTML).Return(nil, nil).Once()
	env.store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(latest, nil).Once()
	env.store.EXPECT().CreateVersion(mock.Anything, mock.MatchedBy(func(version domain.Version) bool {
		return version.Source == domain.VersionSourceRepair
	})).Return(nil)

	body := strings.NewReader(`{"error":"не работает прыжок"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/fix", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestRateGameReturnsAggregate(t *testing.T) {
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
	env.store.EXPECT().UpsertRating(mock.Anything, mock.MatchedBy(func(rating domain.Rating) bool {
		return rating.Value == domain.Rating5
	})).Return(nil)
	env.store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{
		Average: 4.5,
		Total:   4,
		Counts:  [5]int{0, 0, 0, 2, 2},
	}, nil)

	body := strings.NewReader(`{"value":5}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/rate", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.RateGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Rating != 5 || output.Aggregate.Average != 4.5 {
		t.Errorf("output = %+v, want rating 5 and average 4.5", output)
	}
}

func TestRateGameRejectsUnknownValue(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)

	body := strings.NewReader(`{"value":"awesome"}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/rate", body))
	rec := env.do(req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRecordPlaytimeReturnsTotal(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().InsertPlaySession(mock.Anything, mock.MatchedBy(func(session domain.PlaySession) bool {
		return session.GameID == gameID && session.Seconds == 120
	})).Return(nil)
	env.store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{Seconds: 900}, nil)

	body := strings.NewReader(`{"seconds":120}`)
	req := env.authorized(httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/playtime", body))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.RecordPlaytimeOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.TotalSeconds != 900 {
		t.Errorf("total seconds = %d, want 900", output.TotalSeconds)
	}
}

func TestGameRouteRejectsBrokenID(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	for _, tt := range []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/games/not-a-uuid"},
		{method: http.MethodPost, path: "/games/not-a-uuid/patch"},
		{method: http.MethodPost, path: "/games/not-a-uuid/rate"},
	} {
		env.expectSession(userID)

		req := env.authorized(httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`)))
		rec := env.do(req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, http.StatusBadRequest)
		}
	}
}
