-- +goose Up
ALTER TABLE games ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'private';
ALTER TABLE games ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE games DROP CONSTRAINT IF EXISTS games_visibility_check;
ALTER TABLE games ADD CONSTRAINT games_visibility_check CHECK (visibility IN ('public', 'private'));

UPDATE games SET visibility = 'public' WHERE status = 'ready' AND visibility = 'private';

DROP INDEX IF EXISTS games_owner_deleted_idx;
CREATE INDEX games_owner_deleted_idx ON games (owner_id, deleted_at);

DROP INDEX IF EXISTS games_catalog_idx;
CREATE INDEX games_catalog_idx ON games (created_at DESC)
    WHERE status = 'ready' AND visibility = 'public' AND deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS games_catalog_idx;
DROP INDEX IF EXISTS games_owner_deleted_idx;
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_visibility_check;
ALTER TABLE games DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE games DROP COLUMN IF EXISTS visibility;
