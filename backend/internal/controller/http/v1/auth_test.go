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

const (
	testEmail    = "user@example.com"
	testPassword = "password123"
)

func TestRegisterUserSetsSessionCookie(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.auth.EXPECT().HashPassword(testPassword).Return("hash", nil)
	env.auth.EXPECT().GenerateToken().Return(sessionToken, nil)
	env.store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(nil)
	env.store.EXPECT().InsertTokenOp(mock.Anything, mock.Anything).Return(nil)
	env.store.EXPECT().CreateSession(mock.Anything, sessionToken, mock.Anything, mock.Anything).Return(nil)

	body := strings.NewReader(`{"email":"` + testEmail + `","password":"` + testPassword + `"}`)
	rec := env.do(httptest.NewRequest(http.MethodPost, "/auth/register", body))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var output dto.RegisterUserOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.User.Email != testEmail {
		t.Errorf("email = %q, want %q", output.User.Email, testEmail)
	}

	if output.User.Balance != 1_000_000 {
		t.Errorf("balance = %d, want %d", output.User.Balance, int64(1_000_000))
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != sessionToken {
		t.Fatalf("cookies = %+v, want session cookie", cookies)
	}
}

func TestRegisterUserRejectsInvalidBody(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	rec := env.do(httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{"email":"bad"}`)))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestRegisterUserReturnsConflictOnTakenEmail(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.auth.EXPECT().HashPassword(testPassword).Return("hash", nil)
	env.auth.EXPECT().GenerateToken().Return(sessionToken, nil)
	env.store.EXPECT().CreateUser(mock.Anything, mock.Anything).Return(domain.ErrEmailTaken)

	body := strings.NewReader(`{"email":"` + testEmail + `","password":"` + testPassword + `"}`)
	rec := env.do(httptest.NewRequest(http.MethodPost, "/auth/register", body))

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestLoginUserSetsSessionCookie(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.store.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(domain.User{
		ID:           userID,
		Email:        domain.Email(testEmail),
		PasswordHash: domain.PasswordHash("hash"),
	}, nil)
	env.auth.EXPECT().CheckPassword("hash", testPassword).Return(nil)
	env.auth.EXPECT().GenerateToken().Return(sessionToken, nil)
	env.store.EXPECT().CreateSession(mock.Anything, sessionToken, userID, mock.Anything).Return(nil)

	body := strings.NewReader(`{"email":"` + testEmail + `","password":"` + testPassword + `"}`)
	rec := env.do(httptest.NewRequest(http.MethodPost, "/auth/login", body))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != sessionToken {
		t.Fatalf("cookies = %+v, want session cookie", cookies)
	}
}

func TestLoginUserRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.store.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(domain.User{
		ID:           uuid.New(),
		Email:        domain.Email(testEmail),
		PasswordHash: domain.PasswordHash("hash"),
	}, nil)
	env.auth.EXPECT().CheckPassword("hash", "wrong").Return(domain.ErrUnauthorized)

	body := strings.NewReader(`{"email":"` + testEmail + `","password":"wrong"}`)
	rec := env.do(httptest.NewRequest(http.MethodPost, "/auth/login", body))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestLoginUserRejectsUnknownEmail(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.store.EXPECT().GetUserByEmail(mock.Anything, testEmail).Return(domain.User{}, domain.ErrNotFound)

	body := strings.NewReader(`{"email":"` + testEmail + `","password":"` + testPassword + `"}`)
	rec := env.do(httptest.NewRequest(http.MethodPost, "/auth/login", body))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestLogoutUserClearsCookie(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.store.EXPECT().DeleteSession(mock.Anything, sessionToken).Return(nil)

	req := env.authorized(httptest.NewRequest(http.MethodPost, "/auth/logout", nil))
	rec := env.do(req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("cookies = %+v, want cleared session cookie", cookies)
	}
}

func TestMeReturnsBalance(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetUserByID(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		Email:        domain.Email(testEmail),
		TokenBalance: 777,
	}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/me", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.GetMeOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.User.Balance != 777 {
		t.Errorf("balance = %d, want 777", output.User.Balance)
	}
}

func TestProtectedRouteRejectsAnonymous(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	paths := []struct {
		method string
		path   string
	}{
		{method: http.MethodGet, path: "/me"},
		{method: http.MethodGet, path: "/profile"},
		{method: http.MethodPost, path: "/games"},
		{method: http.MethodGet, path: "/games/" + uuid.NewString()},
		{method: http.MethodPost, path: "/games/" + uuid.NewString() + "/patch"},
	}

	for _, tt := range paths {
		rec := env.do(httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{}`)))

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, http.StatusUnauthorized)
		}
	}
}

func TestProtectedRouteRejectsUnknownSession(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.store.EXPECT().GetSessionUserID(mock.Anything, sessionToken).Return(uuid.Nil, domain.ErrNotFound)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/me", nil))
	rec := env.do(req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}
