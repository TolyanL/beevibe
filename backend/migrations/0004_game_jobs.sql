-- +goose Up
CREATE TABLE IF NOT EXISTS game_jobs (
    id          UUID PRIMARY KEY,
    game_id     UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    owner_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    kind        TEXT        NOT NULL,
    instruction TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT game_jobs_kind_check CHECK (kind IN ('patch', 'fix')),
    CONSTRAINT game_jobs_game_kind_unique UNIQUE (game_id, kind)
);

CREATE INDEX IF NOT EXISTS game_jobs_created_idx ON game_jobs (created_at);

-- +goose Down
DROP TABLE IF EXISTS game_jobs;
