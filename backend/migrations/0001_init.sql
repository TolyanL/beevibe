-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    token_balance BIGINT      NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_token_balance_non_negative CHECK (token_balance >= 0)
);

CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT        PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at_idx ON sessions (expires_at);

CREATE TABLE IF NOT EXISTS games (
    id          UUID PRIMARY KEY,
    owner_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    prompt      TEXT        NOT NULL,
    params_json JSONB       NOT NULL DEFAULT '{}'::JSONB,
    spec_json   JSONB,
    config_json JSONB,
    status      TEXT        NOT NULL,
    fail_reason TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS games_owner_created_idx ON games (owner_id, created_at DESC);
CREATE INDEX IF NOT EXISTS games_status_idx ON games (status);

CREATE TABLE IF NOT EXISTS versions (
    id             UUID PRIMARY KEY,
    game_id        UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    number         INT         NOT NULL,
    html           TEXT        NOT NULL,
    source         TEXT        NOT NULL,
    note           TEXT        NOT NULL DEFAULT '',
    tokens_cost    INT         NOT NULL DEFAULT 0,
    model_calls    INT         NOT NULL DEFAULT 0,
    prompt_version TEXT        NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT versions_game_number_unique UNIQUE (game_id, number)
);

CREATE INDEX IF NOT EXISTS versions_game_idx ON versions (game_id);

CREATE TABLE IF NOT EXISTS token_ops (
    id         UUID PRIMARY KEY,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    game_id    UUID        REFERENCES games (id) ON DELETE SET NULL,
    delta      BIGINT      NOT NULL,
    reason     TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS token_ops_user_idx ON token_ops (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ratings (
    game_id    UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    value      TEXT        NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (game_id, user_id),
    CONSTRAINT ratings_value_check CHECK (value IN ('cool', 'bad'))
);

CREATE INDEX IF NOT EXISTS ratings_game_idx ON ratings (game_id);

CREATE TABLE IF NOT EXISTS play_sessions (
    id         UUID PRIMARY KEY,
    game_id    UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    user_id    UUID        REFERENCES users (id) ON DELETE SET NULL,
    seconds    INT         NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT play_sessions_seconds_check CHECK (seconds >= 0)
);

CREATE INDEX IF NOT EXISTS play_sessions_game_idx ON play_sessions (game_id);

CREATE TABLE IF NOT EXISTS gen_events (
    game_id    UUID        NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    seq        BIGINT      NOT NULL,
    type       TEXT        NOT NULL,
    payload    JSONB       NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (game_id, seq)
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    key        TEXT        NOT NULL,
    operation  TEXT        NOT NULL,
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    result_id  TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (key, operation)
);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys;
DROP TABLE IF EXISTS gen_events;
DROP TABLE IF EXISTS play_sessions;
DROP TABLE IF EXISTS ratings;
DROP TABLE IF EXISTS token_ops;
DROP TABLE IF EXISTS versions;
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
