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

func TestGetGameReturnsNotFound(t *testing.T) {
	t.Parallel()

	router, store, _ := newTestRouter(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetSessionUserID(mock.Anything, "session-token").Return(userID, nil)
	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{}, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/games/"+gameID.String(), nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestPatchGameReturnsForbidden(t *testing.T) {
	t.Parallel()

	router, store, _ := newTestRouter(t)

	userID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetSessionUserID(mock.Anything, "session-token").Return(userID, nil)
	store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Status:  domain.StatusReady,
	}, nil)

	body := strings.NewReader(`{"text":"добавь уровень"}`)
	req := httptest.NewRequest(http.MethodPost, "/games/"+gameID.String()+"/patch", body)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestCreateGameReturnsConflictWhenNotEnoughTokens(t *testing.T) {
	t.Parallel()

	router, store, _ := newTestRouter(t)

	userID := uuid.New()

	store.EXPECT().GetSessionUserID(mock.Anything, "session-token").Return(userID, nil)
	store.EXPECT().GetUserByIDForUpdate(mock.Anything, userID).Return(domain.User{
		ID:           userID,
		TokenBalance: 10,
	}, nil)

	body := strings.NewReader(`{"prompt":"змейка"}`)
	req := httptest.NewRequest(http.MethodPost, "/games", body)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
}

func TestLogoutUserClearsCookie(t *testing.T) {
	t.Parallel()

	router, store, _ := newTestRouter(t)

	store.EXPECT().DeleteSession(mock.Anything, "session-token").Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("cookies = %+v, want cleared session cookie", cookies)
	}
}
