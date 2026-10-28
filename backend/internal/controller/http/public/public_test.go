package public

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/usecase"
	gamemocks "github.com/TolyanL/beevibe/internal/usecase/mocks/game"
	llmmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/llm"
	outboxmocks "github.com/TolyanL/beevibe/internal/usecase/mocks/outbox"
	usecasemocks "github.com/TolyanL/beevibe/internal/usecase/mocks/usecase"
	usermocks "github.com/TolyanL/beevibe/internal/usecase/mocks/user"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

func TestMain(m *testing.M) {
	transaction.IsUnitTest = true

	os.Exit(m.Run())
}

const gameHTML = `<!DOCTYPE html><html lang="ru"><body><canvas id="game"></canvas></body></html>`

func newTestRouter(t *testing.T) (http.Handler, *usecasemocks.MockStore) {
	t.Helper()

	store := usecasemocks.NewMockStore(t)

	uc := usecase.New(
		store,
		llmmocks.NewMockLLM(t),
		gamemocks.NewMockValidator(t),
		usermocks.NewMockAuth(t),
		outboxmocks.NewMockKafka(t),
		usecase.Config{},
	)

	return NewRouter(New(uc), nil), store
}

func TestGetGameServesHTML(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPublic,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 1,
		HTML:   gameHTML,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/g/zmeyka-1a2b", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("content type = %q, want text/html", contentType)
	}

	if rec.Body.String() != gameHTML {
		t.Errorf("body = %q, want %q", rec.Body.String(), gameHTML)
	}
}

func TestGetGameNotFound(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPublic,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{}, domain.ErrVersionNotFound)

	req := httptest.NewRequest(http.MethodGet, "/g/zmeyka-1a2b", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetGameTreatsUnknownSlugAsNotFound(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	store.EXPECT().GetGameBySlug(mock.Anything, "unknown-slug").Return(domain.Game{}, domain.ErrNotFound)

	req := httptest.NewRequest(http.MethodGet, "/g/unknown-slug", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestDownloadGameSetsAttachment(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPublic,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 1,
		HTML:   gameHTML,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/g/zmeyka-1a2b/download", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if disposition := rec.Header().Get("Content-Disposition"); disposition != `attachment; filename="game.html"` {
		t.Errorf("content disposition = %q, want attachment", disposition)
	}
}

func TestGetGameServesPrivateGameToItsOwner(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	ownerID := uuid.New()
	gameID := uuid.New()

	store.EXPECT().GetSessionUserID(mock.Anything, "session-token").Return(ownerID, nil).Maybe()
	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:         gameID,
		OwnerID:    ownerID,
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPrivate,
	}, nil)
	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 1,
		HTML:   gameHTML,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/g/zmeyka-1a2b", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})

	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if !strings.Contains(rec.Body.String(), "<canvas") {
		t.Errorf("body = %q, want the game html", rec.Body.String())
	}
}

func TestGetGameHidesPrivateGameFromGuests(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	gameID := uuid.New()

	store.EXPECT().GetGameBySlug(mock.Anything, "zmeyka-1a2b").Return(domain.Game{
		ID:         gameID,
		OwnerID:    uuid.New(),
		Status:     domain.StatusReady,
		Visibility: domain.VisibilityPrivate,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/g/zmeyka-1a2b", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
