-- +goose Up
CREATE TABLE IF NOT EXISTS game_artifacts (
    game_id    UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    kind       TEXT        NOT NULL,
    content    TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (game_id, kind)
);

-- +goose Down
DROP TABLE IF EXISTS game_artifacts;
