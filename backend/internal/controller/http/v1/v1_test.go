package v1

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

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

const (
	sessionToken = "session-token"
	gameHTML     = `<!DOCTYPE html><html lang="ru"><body><canvas id="game"></canvas></body></html>`
)

func TestMain(m *testing.M) {
	transaction.IsUnitTest = true

	os.Exit(m.Run())
}

type testEnv struct {
	router    http.Handler
	store     *usecasemocks.MockStore
	auth      *usermocks.MockAuth
	llm       *llmmocks.MockLLM
	validator *gamemocks.MockValidator
	kafka     *outboxmocks.MockKafka
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	env := &testEnv{
		store:     usecasemocks.NewMockStore(t),
		auth:      usermocks.NewMockAuth(t),
		llm:       llmmocks.NewMockLLM(t),
		validator: gamemocks.NewMockValidator(t),
		kafka:     outboxmocks.NewMockKafka(t),
	}

	uc := usecase.New(
		env.store,
		env.llm,
		env.validator,
		env.auth,
		env.kafka,
		usecase.Config{
			Worker: usecase.Worker{
				Timeout:         60 * time.Second,
				MaxCallsPerGame: 12,
				MaxFixAttempts:  2,
			},
			Tokens: usecase.Tokens{
				GrantSignup:   1_000_000,
				TopupMin:      10_000,
				TopupMax:      1_000_000,
				TopupDailyCap: 5_000_000,
				DemoTopup:     true,
				PriceGenerate: 320_000,
				PricePatch:    40_000,
				PriceMeta:     10_000,
			},
		},
	)

	env.router = NewRouter(New(uc))

	return env
}

func (e *testEnv) authorized(req *http.Request) *http.Request {
	req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})

	return req
}

func (e *testEnv) expectSession(userID uuid.UUID) {
	e.store.EXPECT().GetSessionUserID(mock.Anything, sessionToken).Return(userID, nil).Once()
}

// gamePath отдаёт slug-адрес игры и учит middleware резолвить его в нужный id.
func (e *testEnv) gamePath(gameID uuid.UUID) string {
	slug := gameID.String()

	e.store.EXPECT().GetGameBySlug(mock.Anything, slug).Return(domain.Game{ID: gameID}, nil).Maybe()

	return "/games/" + slug
}

func (e *testEnv) do(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()

	e.router.ServeHTTP(rec, req)

	return rec
}
