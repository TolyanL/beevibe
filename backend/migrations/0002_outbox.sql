-- +goose Up
CREATE TABLE IF NOT EXISTS outbox (
    id           UUID PRIMARY KEY,
    topic        TEXT        NOT NULL,
    key          TEXT        NOT NULL,
    payload      JSONB       NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    attempts     INT         NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS outbox_unpublished_idx ON outbox (created_at) WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS game_event_log (
    id          UUID PRIMARY KEY,
    event_id    UUID        NOT NULL UNIQUE,
    topic       TEXT        NOT NULL,
    key         TEXT        NOT NULL,
    event_type  TEXT        NOT NULL,
    payload     JSONB       NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS game_event_log_type_idx ON game_event_log (event_type);

-- +goose Down
DROP TABLE IF EXISTS game_event_log;
DROP TABLE IF EXISTS outbox;
