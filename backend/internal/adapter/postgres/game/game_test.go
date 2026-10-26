package game

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
	"github.com/TolyanL/beevibe/internal/domain"
)

const (
	gameColumns    = 10
	versionColumns = 10
)

func TestGetGameScansRowWithJSON(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	ownerID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args(gameID)).Return(pgtest.Row(t,
		gameID,
		ownerID,
		"змейка",
		[]byte(`{"genre":"snake","tempo":6,"map_size":"small","style":"neon"}`),
		[]byte(`{"genre":"snake"}`),
		[]byte(`{"tempo":6}`),
		"ready",
		"",
		createdAt,
		createdAt,
	))

	game, err := postgres.GetGame(ctx, gameID)
	if err != nil {
		t.Fatalf("GetGame() error = %v, want nil", err)
	}

	if game.ID != gameID || game.OwnerID != ownerID || game.Status != domain.StatusReady {
		t.Errorf("game = %+v, want id %s owner %s status ready", game, gameID, ownerID)
	}

	if game.Params.Genre != domain.GenreSnake || game.Params.Tempo != 6 || game.Params.Style != domain.StyleNeon {
		t.Errorf("params = %+v, want snake 6 neon", game.Params)
	}

	if game.Spec == nil || game.Spec.Genre != "snake" {
		t.Errorf("spec = %+v, want genre snake", game.Spec)
	}

	if game.Config == nil || game.Config.Tempo != 6 {
		t.Errorf("config = %+v, want tempo 6", game.Config)
	}
}

func TestGetGameKeepsNullablesEmpty(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	now := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Row(t,
		gameID, uuid.New(), "змейка", nil, nil, nil, "queued", "", now, now,
	))

	game, err := postgres.GetGame(ctx, gameID)
	if err != nil {
		t.Fatalf("GetGame() error = %v, want nil", err)
	}

	if game.Spec != nil || game.Config != nil {
		t.Errorf("spec = %v config = %v, want nil", game.Spec, game.Config)
	}
}

func TestGetGameMapsNoRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.RowError(t, pgx.ErrNoRows, gameColumns))

	_, err := postgres.GetGame(ctx, uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetGame() error = %v, want ErrNotFound", err)
	}
}

func TestGetGameWrapsBrokenJSON(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	now := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).Return(pgtest.Row(t,
		gameID, uuid.New(), "змейка", []byte(`{broken`), nil, nil, "queued", "", now, now,
	))

	if _, err := postgres.GetGame(ctx, gameID); err == nil {
		t.Fatal("GetGame() error = nil, want error")
	}
}

func TestClaimQueuedGameReturnsClaimedGame(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	ownerID := uuid.New()
	now := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything).Return(pgtest.Row(t,
		gameID, ownerID, "змейка", nil, nil, nil, "analyzing", "", now, now,
	))

	game, err := postgres.ClaimQueuedGame(ctx)
	if err != nil {
		t.Fatalf("ClaimQueuedGame() error = %v, want nil", err)
	}

	if game.ID != gameID || game.Status != domain.StatusAnalyzing {
		t.Errorf("game = %+v, want id %s status analyzing", game, gameID)
	}
}

func TestClaimQueuedGameReturnsNotFoundWhenQueueIsEmpty(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything).
		Return(pgtest.RowError(t, pgx.ErrNoRows, gameColumns))

	_, err := postgres.ClaimQueuedGame(ctx)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ClaimQueuedGame() error = %v, want ErrNotFound", err)
	}
}

func TestListVersionsScansRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(gameID)).Return(pgtest.Rows(t,
		[]any{uuid.New(), gameID, 1, "<html>v1</html>", "generate", "", 100, 6, "v1", createdAt},
		[]any{uuid.New(), gameID, 2, "<html>v2</html>", "patch", "кнопки", 40, 1, "v1", createdAt},
	), nil)

	versions, err := postgres.ListVersions(ctx, gameID)
	if err != nil {
		t.Fatalf("ListVersions() error = %v, want nil", err)
	}

	if len(versions) != 2 {
		t.Fatalf("versions = %d, want 2", len(versions))
	}

	if versions[0].Number != 1 || versions[0].Source != domain.VersionSourceGenerate || versions[0].HTML != "<html>v1</html>" {
		t.Errorf("first version = %+v, want generate 1", versions[0])
	}

	if versions[1].Number != 2 || versions[1].Source != domain.VersionSourcePatch || versions[1].Note != "кнопки" {
		t.Errorf("second version = %+v, want patch 2 with note", versions[1])
	}
}

func TestListVersionsWrapsIterationError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.RowsError(t, errors.New("connection reset")), nil)

	if _, err := postgres.ListVersions(ctx, uuid.New()); err == nil {
		t.Fatal("ListVersions() error = nil, want error")
	}
}

func TestInsertGenEventStoresEvent(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	event := domain.GenEvent{
		GameID:    uuid.New(),
		Type:      domain.GenEventStage,
		Payload:   []byte(`{"stage":"analyzing"}`),
		CreatedAt: time.Now().UTC(),
	}

	db.EXPECT().QueryRow(mock.Anything, mock.Anything,
		pgtest.Args(event.GameID, string(event.Type), string(event.Payload), mock.Anything)).
		Return(pgtest.Row(t, int64(1)))

	if err := postgres.InsertGenEvent(ctx, event); err != nil {
		t.Fatalf("InsertGenEvent() error = %v, want nil", err)
	}
}

func TestInsertGenEventRetriesOnSequenceCollision(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	event := domain.GenEvent{GameID: uuid.New(), Type: domain.GenEventStage, Payload: []byte(`{}`)}

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.RowError(t, pgtest.UniqueViolation("gen_events_pkey"), 1)).Once()
	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.Row(t, int64(2))).Once()

	if err := postgres.InsertGenEvent(ctx, event); err != nil {
		t.Fatalf("InsertGenEvent() error = %v, want nil", err)
	}
}

func TestInsertGenEventGivesUpAfterAttempts(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	event := domain.GenEvent{GameID: uuid.New(), Type: domain.GenEventStage, Payload: []byte(`{}`)}
	collision := pgtest.RowError(t, pgtest.UniqueViolation("gen_events_pkey"), 1)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(collision).Times(genEventInsertAttempts)

	if err := postgres.InsertGenEvent(ctx, event); err == nil {
		t.Fatal("InsertGenEvent() error = nil, want error")
	}
}

func TestCreateVersionInsertsRow(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	version := domain.Version{
		ID:         uuid.New(),
		GameID:     uuid.New(),
		Number:     2,
		HTML:       "<html></html>",
		Source:     domain.VersionSourcePatch,
		Note:       "кнопки",
		TokensCost: 40,
		ModelCalls: 1,
		CreatedAt:  time.Now().UTC(),
	}

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(
		version.ID, version.GameID, version.Number, version.HTML, string(version.Source),
		version.Note, version.TokensCost, version.ModelCalls, mock.Anything, mock.Anything,
	)).Return(pgtest.Tag(1), nil)

	if err := postgres.CreateVersion(ctx, version); err != nil {
		t.Fatalf("CreateVersion() error = %v, want nil", err)
	}
}

func TestGetLatestVersionReturnsVersion(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args(gameID)).Return(pgtest.Row(t,
		uuid.New(), gameID, 3, "<html></html>", "config", "config update", 0, 0, "v1", createdAt,
	))

	version, err := postgres.GetLatestVersion(ctx, gameID)
	if err != nil {
		t.Fatalf("GetLatestVersion() error = %v, want nil", err)
	}

	if version.Number != 3 || version.Source != domain.VersionSourceConfig {
		t.Errorf("version = %+v, want config 3", version)
	}
}

func TestGetLatestVersionMapsNoRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, mock.Anything).
		Return(pgtest.RowError(t, pgx.ErrNoRows, versionColumns))

	_, err := postgres.GetLatestVersion(ctx, uuid.New())
	if !errors.Is(err, domain.ErrVersionNotFound) {
		t.Fatalf("GetLatestVersion() error = %v, want ErrVersionNotFound", err)
	}
}

func TestGetRatingAggregateComputesAverage(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args(gameID)).Return(pgtest.Row(t, 4.5, 4, 0, 0, 0, 2, 2))

	aggregate, err := postgres.GetRatingAggregate(ctx, gameID)
	if err != nil {
		t.Fatalf("GetRatingAggregate() error = %v, want nil", err)
	}

	if aggregate.Total != 4 || aggregate.Average != 4.5 || aggregate.Counts != [5]int{0, 0, 0, 2, 2} {
		t.Errorf("aggregate = %+v, want total 4, average 4.5, counts 0/0/0/2/2", aggregate)
	}
}

func TestGetPlaytimeAggregateComputesHours(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()

	db.EXPECT().QueryRow(mock.Anything, mock.Anything, pgtest.Args(gameID)).Return(pgtest.Row(t, 2, 1800))

	aggregate, err := postgres.GetPlaytimeAggregate(ctx, gameID)
	if err != nil {
		t.Fatalf("GetPlaytimeAggregate() error = %v, want nil", err)
	}

	if aggregate.Sessions != 2 || aggregate.Seconds != 1800 || aggregate.Hours != 0.5 {
		t.Errorf("aggregate = %+v, want 2 sessions 1800s 0.5h", aggregate)
	}
}

func TestListGenEventsScansRowsAfterSequence(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()
	createdAt := time.Now().UTC()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(gameID, int64(2))).Return(pgtest.Rows(t,
		[]any{gameID, int64(3), "stage", []byte(`{"stage":"coding"}`), createdAt},
	), nil)

	events, err := postgres.ListGenEvents(ctx, gameID, 2)
	if err != nil {
		t.Fatalf("ListGenEvents() error = %v, want nil", err)
	}

	if len(events) != 1 || events[0].Seq != 3 || events[0].Type != domain.GenEventStage {
		t.Errorf("events = %+v, want one stage event with seq 3", events)
	}
}
