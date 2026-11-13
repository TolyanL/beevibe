package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/dto"
)

func TestListUserGenerationsReturnsPage(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().ListUserGenerations(mock.Anything, userID, dto.ListUserGenerationsInput{Limit: 10, Offset: 5}).
		Return([]dto.ProfileGeneration{
			{
				GameID:        gameID,
				GameSlug:      "zmeyka-1a2b",
				GameTitle:     "Змейка",
				GamePrompt:    "змейка",
				GameStatus:    "ready",
				VersionNumber: 3,
				Source:        "patch",
				Note:          "плавнее",
				TokensCost:    40,
				ModelCalls:    1,
				CreatedAt:     time.Now().UTC(),
			},
		}, 12, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/me/generations?limit=10&offset=5", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.ListUserGenerationsOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if len(output.Items) != 1 || output.Total != 12 {
		t.Fatalf("output = %+v, want one generation with total 12", output)
	}

	if output.Items[0].GameID != gameID || output.Items[0].VersionNumber != 3 || output.Items[0].Source != "patch" {
		t.Errorf("item = %+v, want version 3 patch of %s", output.Items[0], gameID)
	}
}

func TestListUserGenerationsRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	env.expectSession(uuid.New())

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/me/generations?limit=1000", nil))
	rec := env.do(req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestListUserGenerationsRequiresAuth(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/me/generations", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
