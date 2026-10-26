package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/domain"
	"github.com/TolyanL/beevibe/internal/dto"
)

func TestGetGamePublicCardWithoutAuth(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	gameID := uuid.New()
	createdAt := time.Now().UTC()

	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:       gameID,
		OwnerID:  uuid.New(),
		Prompt:   "змейка в неоне",
		Params:   domain.GameParams{Genre: domain.GenreSnake, Style: domain.StyleNeon, MapSize: domain.MapSizeMedium},
		Status:   domain.StatusReady,
		CreatedAt: createdAt,
	}, nil)
	env.store.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(domain.User{
		ID:          uuid.New(),
		Email:       "author@example.com",
		DisplayName: "Автор",
	}, nil).Maybe()
	env.store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{Average: 4.5, Total: 2, Counts: [5]int{0, 0, 0, 1, 1}}, nil)
	env.store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{Sessions: 3, Seconds: 5400, Hours: 1.5}, nil)

	rec := env.do(httptest.NewRequest(http.MethodGet, "/games/"+gameID.String()+"/public", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.GetGamePublicCardOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.ID != gameID || output.Prompt != "змейка в неоне" || output.Genre != "snake" || output.Style != "neon" {
		t.Errorf("card = %+v, want snake/neon card of game %s", output, gameID)
	}

	if output.Rating.Average != 4.5 || output.Playtime.Hours != 1.5 {
		t.Errorf("aggregates = %+v %+v, want rating 4.5 and playtime 1.5h", output.Rating, output.Playtime)
	}

	if output.MyRating != 0 {
		t.Errorf("myRating = %d, want 0 for anonymous viewer", output.MyRating)
	}
}

func TestGetGamePublicCardWithAuthIncludesMyRating(t *testing.T) {
	t.Parallel()

	env := newTestEnv(t)

	userID := uuid.New()
	gameID := uuid.New()

	env.expectSession(userID)
	env.store.EXPECT().GetGame(mock.Anything, gameID).Return(domain.Game{
		ID:      gameID,
		OwnerID: uuid.New(),
		Prompt:  "змейка",
		Status:  domain.StatusReady,
	}, nil)
	env.store.EXPECT().GetUserByID(mock.Anything, mock.Anything).Return(domain.User{
		ID:          uuid.New(),
		Email:       "author@example.com",
		DisplayName: "Автор",
	}, nil).Maybe()
	env.store.EXPECT().GetRatingAggregate(mock.Anything, gameID).Return(dto.RatingAggregate{}, nil)
	env.store.EXPECT().GetPlaytimeAggregate(mock.Anything, gameID).Return(dto.PlaytimeAggregate{}, nil)
	env.store.EXPECT().GetRating(mock.Anything, gameID, userID).Return(domain.Rating{Value: domain.Rating4}, nil)

	req := env.authorized(httptest.NewRequest(http.MethodGet, "/games/"+gameID.String()+"/public", nil))
	rec := env.do(req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var output dto.GetGamePublicCardOutput
	if err := json.NewDecoder(rec.Body).Decode(&output); err != nil {
		t.Fatalf("decode output: %v", err)
	}

	if output.MyRating != 4 {
		t.Errorf("myRating = %d, want 4", output.MyRating)
	}
}
