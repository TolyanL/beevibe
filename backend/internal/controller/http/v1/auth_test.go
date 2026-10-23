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
	"github.com/TolyanL/beevibe/internal/usecase"
	gamemocks "github.com/TolyanL/beevibe/internal/usecase/mocks/game"
	llmmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/llm"
	outboxmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/outbox"
	usecasemocks "github.com/TolyanL/beevibe/internal/usecase/mocks/usecase"
	usermocks "github.com/TolyanL/beevibe/internal/usecase/mocks/user"
)

func newTestRouter(t *testing.T) (http.Handler, *usecasemocks.MockStore, *usermocks.MockAuth) {
	t.Helper()

	store := usecasemocks.NewMockStore(t)
	auth := usermocks.NewMockAuth(t)

	uc := usecase.New(
		store,
		llmmocks.NewMockLLM(t),
		gamemocks.NewMockValidator(t),
		auth,
		outboxmocks.NewMockKafka(t),
		usecase.Config{
			Tokens: usecase.Tokens{
				GrantSignup:   1_000_000,
				GrantTopup:    500_000,
				PriceGenerate: 320_000,
				PricePatch:    40_000,
			},
		},
	)

	return NewRouter(New(uc)), store, auth
}

func TestRegisterUserSetsSessionCookie(t *testing.T) {
	t.Parallel()

	router, store, auth := newTestRouter(t)

	auth.EXPECT().HashPassword("password123").Return("hash", nil)
	auth.EXPECT().GenerateToken().Return("session-token", nil)
	store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().CreateSession(mock.Anything, "session-token", mock.Anything, mock.Anything).Return(nil)

	body := strings.NewReader(`{"email":"user@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	cookie := rec.Result().Cookies()
	if len(cookie) != 1 || cookie[0].Name != "session" || cookie[0].Value != "session-token" {
		t.Fatalf("cookies = %+v, want session cookie", cookie)
	}
}

func TestRegisterUserRejectsInvalidBody(t *testing.T) {
	t.Parallel()

	router, _, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":"bad"}`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestLoginUserSetsSessionCookie(t *testing.T) {
	t.Parallel()

	router, store, auth := newTestRouter(t)

	userID := uuid.New()

	store.EXPECT().GetUserByEmail(mock.Anything, "user@example.com").Return(domain.User{
		ID:           userID,
		Email:        domain.Email("user@example.com"),
		PasswordHash: domain.PasswordHash("hash"),
	}, nil)
	auth.EXPECT().CheckPassword("hash", "password123").Return(nil)
	auth.EXPECT().GenerateToken().Return("session-token", nil)
	store.EXPECT().CreateSession(mock.Anything, "session-token", userID, mock.Anything).Return(nil)

	body := strings.NewReader(`{"email":"user@example.com","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	cookie := rec.Result().Cookies()
	if len(cookie) != 1 || cookie[0].Value != "session-token" {
		t.Fatalf("cookies = %+v, want session cookie", cookie)
	}
}

func TestLoginUserRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	router, store, auth := newTestRouter(t)

	store.EXPECT().GetUserByEmail(mock.Anything, "user@example.com").Return(domain.User{
		ID:           uuid.New(),
		Email:        domain.Email("user@example.com"),
		PasswordHash: domain.PasswordHash("hash"),
	}, nil)
	auth.EXPECT().CheckPassword("hash", "wrong").Return(domain.ErrUnauthorized)

	body := strings.NewReader(`{"email":"user@example.com","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestProtectedRouteRejectsAnonymous(t *testing.T) {
	t.Parallel()

	router, _, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestCreateGameReturnsGameID(t *testing.T) {
	t.Parallel()

	router, store, _ := newTestRouter(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetSessionUserID(mock.Anything, "session-token").Return(userID, nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 1_000_000,
	}, nil)
	store.EXPECT().UpdateUser(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertTokenOp(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().CreateGame(mock.Anything, mock.Anything).Return(nil)
	store.EXPECT().InsertOutboxEvent(mock.Anything, mock.Anything).Return(nil)

	body := strings.NewReader(`{"prompt":"змейка","genre":"snake"}`)
	req := httptest.NewRequest(http.MethodPost, "/games", body)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var output dto.CreateGameOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.Status != string(domain.StatusQueued) {
		t.Errorf("status = %q, want %q", output.Status, domain.StatusQueued)
	}

	if output.ID == uuid.Nil || output.ID == gameID {
		t.Errorf("id = %s, want generated uuid", output.ID)
	}
}

func TestHealthReturnsStatus(t *testing.T) {
	t.Parallel()

	router, _, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}
