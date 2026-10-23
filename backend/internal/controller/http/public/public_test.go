package public

import (
	"net/http"
	"net/http/httptest"
	"os"
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

	return NewRouter(New(uc)), store
}

func TestGetGameServesHTML(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	gameID := uuid.New()

	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 1,
		HTML:   gameHTML,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/g/"+gameID.String(), nil)
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

	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{}, domain.ErrVersionNotFound)

	req := httptest.NewRequest(http.MethodGet, "/g/"+gameID.String(), nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestGetGameRejectsBrokenID(t *testing.T) {
	t.Parallel()

	router, _ := newTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/g/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestDownloadGameSetsAttachment(t *testing.T) {
	t.Parallel()

	router, store := newTestRouter(t)

	gameID := uuid.New()

	store.EXPECT().GetLatestVersion(mock.Anything, gameID).Return(domain.Version{
		ID:     uuid.New(),
		GameID: gameID,
		Number: 1,
		HTML:   gameHTML,
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/g/"+gameID.String()+"/download", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	if disposition := rec.Header().Get("Content-Disposition"); disposition != `attachment; filename="game.html"` {
		t.Errorf("content disposition = %q, want attachment", disposition)
	}
}
