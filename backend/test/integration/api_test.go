//go:build integration

package integration

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/TolyanL/beevibe/internal/adapter/auth"
	"github.com/TolyanL/beevibe/internal/adapter/llm"
	"github.com/TolyanL/beevibe/internal/adapter/postgres"
	"github.com/TolyanL/beevibe/internal/adapter/validator"
	"github.com/TolyanL/beevibe/internal/app"
	"github.com/TolyanL/beevibe/internal/controller/http/public"
	"github.com/TolyanL/beevibe/internal/controller/http/v1"
	"github.com/TolyanL/beevibe/internal/usecase"
	"github.com/TolyanL/beevibe/migration"
	pgpool "github.com/TolyanL/beevibe/pkg/postgres"
	"github.com/TolyanL/beevibe/pkg/transaction"
)

const gameHTML = `<!DOCTYPE html>
<html lang="ru">
<head><meta charset="utf-8"><title>Змейка</title><style>body{margin:0;background:#101010}</style></head>
<body>
<canvas id="game" width="300" height="300"></canvas>
<script>
const CONFIG = { tempo: 5, palette: ["#101010", "#33ff66"], mapSize: "medium", difficulty: "normal", controls: "arrows", style: "neon" };
const canvas = document.getElementById('game');
const ctx = canvas.getContext('2d');
let state = 'start';
function start() { state = 'play'; }
function win() { state = 'win'; }
function lose() { state = 'lose'; }
function restart() { state = 'start'; }
canvas.addEventListener('touchstart', function (event) { event.preventDefault(); start(); });
document.addEventListener('keydown', function (event) { if (event.key === ' ') { start(); } });
</script>
</body>
</html>`

const analyzerJSON = `{"genre":"snake","mechanic":"classic snake","controls":"arrows and swipe",` +
	`"win_condition":"fill the board","lose_condition":"hit the wall","palette":["#101010","#33ff66"],` +
	`"map_size":"medium","style":"neon","audience":"everyone","assets":["apple","snake"]}`

type fakeOpenRouter struct {
	server *httptest.Server

	mu    sync.Mutex
	calls []string
}

func newFakeOpenRouter(t *testing.T) *fakeOpenRouter {
	t.Helper()

	f := &fakeOpenRouter{}
	f.server = httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(f.server.Close)

	return f
}

func (f *fakeOpenRouter) handle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		return
	}

	system := ""
	if len(req.Messages) > 0 {
		system = req.Messages[0].Content
	}

	f.mu.Lock()
	f.calls = append(f.calls, system)
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	//nolint:errcheck // тестовый сервер: ошибка кодирования не влияет на asserts
	json.NewEncoder(w).Encode(map[string]any{
		"choices": []map[string]any{{
			"message": map[string]string{"role": "assistant", "content": fakeAnswer(system)},
		}},
		"usage": map[string]int{"prompt_tokens": 100, "completion_tokens": 200},
	})
}

func (f *fakeOpenRouter) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.calls)
}

func fakeAnswer(system string) string {
	switch {
	case strings.Contains(system, "# Analyzer"):
		return analyzerJSON
	case strings.Contains(system, "# Reviewer"):
		return `{"issues":[],"approved":true}`
	case strings.Contains(system, "# Assembler"), strings.Contains(system, "# Patch"), strings.Contains(system, "# Repair"):
		return gameHTML
	default:
		return "ok"
	}
}

type APISuite struct {
	suite.Suite

	server *httptest.Server
	llm    *fakeOpenRouter
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func (s *APISuite) SetupSuite() {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://bee:12345@localhost:5432/bee?sslmode=disable"
	}

	ctx := s.T().Context()

	if err := migration.Up(ctx, dsn); err != nil {
		s.T().Fatalf("migration.Up: %v", err)
	}

	pool, err := pgpool.New(ctx, pgpool.Config{DSN: dsn})
	if err != nil {
		s.T().Fatalf("pgpool.New: %v", err)
	}

	transaction.Init(pool)

	s.llm = newFakeOpenRouter(s.T())

	store := postgres.New(pool)
	llmClient := llm.New(llm.Config{
		APIKey:       "test-key",
		Model:        "test/model",
		BaseURL:      s.llm.server.URL,
		Timeout:      10 * time.Second,
		MaxRetries:   1,
		RetryBackoff: 10 * time.Millisecond,
	})

	uc := usecase.New(store, llmClient, validator.New(validator.Config{}), auth.New(auth.Config{BcryptCost: 4}), usecase.Config{
		Tokens: usecase.Tokens{
			GrantSignup:   1_000_000,
			GrantTopup:    500_000,
			PriceGenerate: 320_000,
			PricePatch:    40_000,
		},
		Limits: usecase.LLMLimits{
			MaxTokensAnalyze:  1000,
			MaxTokensPlan:     1000,
			MaxTokensCode:     2000,
			MaxTokensAssets:   1000,
			MaxTokensReview:   1000,
			MaxTokensAssemble: 2000,
			MaxTokensPatch:    2000,
			MaxTokensRepair:   2000,
		},
		Worker: usecase.Worker{
			Timeout:         60 * time.Second,
			MaxCallsPerGame: 12,
			MaxFixAttempts:  2,
		},
		App: usecase.App{Name: "gamegen", Version: "integration"},
	})

	worker := app.NewWorker(uc, 2)
	worker.Start(ctx)
	s.T().Cleanup(worker.Close)

	root := chi.NewRouter()
	root.Mount("/api", v1.NewRouter(v1.New(uc)))
	root.Mount("/", public.NewRouter(public.New(uc)))

	s.server = httptest.NewServer(root)
	s.T().Cleanup(s.server.Close)
}

func (s *APISuite) register(email string) *http.Client {
	jar, err := cookiejar.New(nil)
	s.Require().NoError(err)

	client := &http.Client{Jar: jar}

	body := fmt.Sprintf(`{"email":%q,"password":"password123"}`, email)

	resp, err := client.Post(s.server.URL+"/api/auth/register", "application/json", strings.NewReader(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	return client
}

func (s *APISuite) TestRegisterAndMe() {
	email := fmt.Sprintf("user-%s@example.com", uuid.NewString())

	client := s.register(email)

	resp, err := client.Get(s.server.URL + "/api/me")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var output struct {
		User struct {
			Email   string `json:"email"`
			Balance int64  `json:"balance"`
		} `json:"user"`
	}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&output))
	s.Require().Equal(email, output.User.Email)
	s.Require().Equal(int64(1_000_000), output.User.Balance)
}

func (s *APISuite) TestRegisterDuplicateEmail() {
	email := fmt.Sprintf("dup-%s@example.com", uuid.NewString())

	s.register(email)

	resp, err := http.Post(s.server.URL+"/api/auth/register", "application/json",
		strings.NewReader(fmt.Sprintf(`{"email":%q,"password":"password123"}`, email)))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusConflict, resp.StatusCode)
}

func (s *APISuite) TestGenerateGameAndStream() {
	client := s.register(fmt.Sprintf("gen-%s@example.com", uuid.NewString()))

	body := `{"prompt":"змейка для бабушки","genre":"snake","tempo":6,"mapSize":"small","style":"neon"}`

	resp, err := client.Post(s.server.URL+"/api/games", "application/json", strings.NewReader(body))
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusCreated, resp.StatusCode)

	var created struct {
		ID     uuid.UUID `json:"id"`
		Status string    `json:"status"`
	}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&created))
	s.Require().Equal("queued", created.Status)

	game := s.waitForReady(client, created.ID)
	s.Require().Equal("ready", game.Status)
	s.Require().NotEmpty(game.Spec)
	s.Require().NotEmpty(game.Config)
	s.Require().Len(game.Versions, 1)
	s.Require().Equal("generate", game.Versions[0].Source)

	publicResp, err := http.Get(s.server.URL + "/g/" + created.ID.String())
	s.Require().NoError(err)
	defer publicResp.Body.Close()

	s.Require().Equal(http.StatusOK, publicResp.StatusCode)
	s.Require().Contains(publicResp.Header.Get("Content-Type"), "text/html")
}

func (s *APISuite) TestStreamReturnsStageAndDoneEvents() {
	client := s.register(fmt.Sprintf("stream-%s@example.com", uuid.NewString()))

	created := s.createGame(client, `{"prompt":"тетрис","genre":"puzzle"}`, "")

	req, err := http.NewRequest(http.MethodGet, s.server.URL+"/api/games/"+created.ID.String()+"/stream", nil)
	s.Require().NoError(err)

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		s.T().Fatalf("stream: status %d, body %s", resp.StatusCode, string(body))
	}

	s.Require().Contains(resp.Header.Get("Content-Type"), "text/event-stream")

	events := s.readEvents(resp.Body)

	s.Require().Contains(events, "event: stage")
	s.Require().Contains(events, "event: done")
}

func (s *APISuite) readEvents(body io.Reader) string {
	scanner := bufio.NewScanner(body)

	var out strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event:") {
			out.WriteString(line)
			out.WriteString("\n")
		}
	}

	return out.String()
}

func (s *APISuite) TestIdempotencyKeyReturnsSameGame() {
	client := s.register(fmt.Sprintf("idem-%s@example.com", uuid.NewString()))

	body := `{"prompt":"кликер","genre":"clicker"}`
	key := uuid.NewString()

	first := s.createGame(client, body, key)
	second := s.createGame(client, body, key)

	s.Require().Equal(first.ID, second.ID)
}

func (s *APISuite) TestPatchAndRateAndCatalog() {
	client := s.register(fmt.Sprintf("patch-%s@example.com", uuid.NewString()))

	created := s.createGame(client, `{"prompt":"платформер","genre":"platformer"}`, "")
	s.waitForReady(client, created.ID)

	balanceBefore := s.balance(client)

	patchResp, err := client.Post(s.server.URL+"/api/games/"+created.ID.String()+"/patch", "application/json",
		strings.NewReader(`{"text":"сделай кнопки больше"}`))
	s.Require().NoError(err)
	defer patchResp.Body.Close()

	s.Require().Equal(http.StatusOK, patchResp.StatusCode)

	var patch struct {
		VersionID uuid.UUID `json:"versionId"`
	}

	s.Require().NoError(json.NewDecoder(patchResp.Body).Decode(&patch))
	s.Require().NotEqual(uuid.Nil, patch.VersionID)

	s.Require().Less(s.balance(client), balanceBefore)

	configReq, err := http.NewRequest(http.MethodPatch, s.server.URL+"/api/games/"+created.ID.String()+"/config",
		strings.NewReader(`{"tempo":9}`))
	s.Require().NoError(err)

	configReq.Header.Set("Content-Type", "application/json")

	configResp, err := client.Do(configReq)
	s.Require().NoError(err)
	defer configResp.Body.Close()

	s.Require().Equal(http.StatusOK, configResp.StatusCode)

	rateResp, err := client.Post(s.server.URL+"/api/games/"+created.ID.String()+"/rate", "application/json",
		strings.NewReader(`{"value":"cool"}`))
	s.Require().NoError(err)
	defer rateResp.Body.Close()

	s.Require().Equal(http.StatusOK, rateResp.StatusCode)

	playResp, err := client.Post(s.server.URL+"/api/games/"+created.ID.String()+"/playtime", "application/json",
		strings.NewReader(`{"seconds":120}`))
	s.Require().NoError(err)
	defer playResp.Body.Close()

	s.Require().Equal(http.StatusOK, playResp.StatusCode)

	catalogResp, err := http.Get(s.server.URL + "/api/catalog?sort=fresh")
	s.Require().NoError(err)
	defer catalogResp.Body.Close()

	s.Require().Equal(http.StatusOK, catalogResp.StatusCode)

	var catalog struct {
		Items []struct {
			ID     uuid.UUID `json:"id"`
			Rating struct {
				CoolPercent float64 `json:"cool_percent"`
			} `json:"rating"`
		} `json:"items"`
	}

	s.Require().NoError(json.NewDecoder(catalogResp.Body).Decode(&catalog))
	s.Require().NotEmpty(catalog.Items)

	found := false

	for _, item := range catalog.Items {
		if item.ID == created.ID {
			found = true

			s.Require().Equal(float64(100), item.Rating.CoolPercent)
		}
	}

	s.Require().True(found, "game is missing in catalog")
}

func (s *APISuite) TestForeignGameIsForbidden() {
	owner := s.register(fmt.Sprintf("owner-%s@example.com", uuid.NewString()))
	other := s.register(fmt.Sprintf("other-%s@example.com", uuid.NewString()))

	created := s.createGame(owner, `{"prompt":"головоломка","genre":"puzzle"}`, "")
	s.waitForReady(owner, created.ID)

	resp, err := other.Get(s.server.URL + "/api/games/" + created.ID.String())
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusForbidden, resp.StatusCode)
}

func (s *APISuite) TestUnauthenticatedIsRejected() {
	resp, err := http.Get(s.server.URL + "/api/me")
	s.Require().NoError(err)
	defer resp.Body.Close()

	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
}

type gameResponse struct {
	Status   string `json:"status"`
	Spec     any    `json:"spec"`
	Config   any    `json:"config"`
	Versions []struct {
		Number int    `json:"number"`
		Source string `json:"source"`
	} `json:"versions"`
}

func (s *APISuite) createGame(client *http.Client, body, idempotencyKey string) struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
} {
	req, err := http.NewRequest(http.MethodPost, s.server.URL+"/api/games", strings.NewReader(body))
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/json")

	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		s.T().Fatalf("create game: status %d, body %s", resp.StatusCode, string(body))
	}

	var created struct {
		ID     uuid.UUID `json:"id"`
		Status string    `json:"status"`
	}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&created))

	return created
}

func (s *APISuite) waitForReady(client *http.Client, gameID uuid.UUID) gameResponse {
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(s.server.URL + "/api/games/" + gameID.String())
		s.Require().NoError(err)

		var game gameResponse

		s.Require().NoError(json.NewDecoder(resp.Body).Decode(&game))

		resp.Body.Close()

		if game.Status == "ready" || game.Status == "failed" {
			return game
		}

		time.Sleep(200 * time.Millisecond)
	}

	s.T().Fatal("game did not reach ready state")

	return gameResponse{}
}

func (s *APISuite) balance(client *http.Client) int64 {
	resp, err := client.Get(s.server.URL + "/api/me")
	s.Require().NoError(err)
	defer resp.Body.Close()

	var output struct {
		User struct {
			Balance int64 `json:"balance"`
		} `json:"user"`
	}

	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&output))

	return output.User.Balance
}
