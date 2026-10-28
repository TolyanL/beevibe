package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/TolyanL/beevibe/internal/adapter/postgres/pgtest"
)

func TestSaveGameArtifactUpserts(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()

	db.EXPECT().Exec(mock.Anything, mock.Anything, pgtest.Args(gameID, "plan", "план", mock.Anything)).
		Return(pgtest.Tag(1), nil)

	if err := postgres.SaveGameArtifact(ctx, gameID, "plan", "план"); err != nil {
		t.Fatalf("SaveGameArtifact() error = %v, want nil", err)
	}
}

func TestListGameArtifactsReadsRows(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	gameID := uuid.New()

	db.EXPECT().Query(mock.Anything, mock.Anything, pgtest.Args(gameID)).Return(pgtest.Rows(t,
		[]any{"plan", "план"},
		[]any{"html", "<html></html>"},
	), nil)

	artifacts, err := postgres.ListGameArtifacts(ctx, gameID)
	if err != nil {
		t.Fatalf("ListGameArtifacts() error = %v, want nil", err)
	}

	if len(artifacts) != 2 || artifacts["html"] != "<html></html>" {
		t.Errorf("artifacts = %v, want plan and html", artifacts)
	}
}

func TestListGameArtifactsWrapsQueryError(t *testing.T) {
	t.Parallel()

	db, ctx := pgtest.DB(t)
	postgres := New(nil)

	db.EXPECT().Query(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("connection reset"))

	if _, err := postgres.ListGameArtifacts(ctx, uuid.New()); err == nil {
		t.Fatal("ListGameArtifacts() error = nil, want error")
	}
}
