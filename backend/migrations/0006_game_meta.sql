-- +goose Up
ALTER TABLE games ADD COLUMN IF NOT EXISTS title TEXT NOT NULL DEFAULT '';
ALTER TABLE games ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';
ALTER TABLE games ADD COLUMN IF NOT EXISTS cover_json JSONB;

UPDATE games SET title = '' WHERE title IS NULL;

ALTER TABLE games DROP CONSTRAINT IF EXISTS game_jobs_kind_check;
ALTER TABLE game_jobs ADD CONSTRAINT game_jobs_kind_check_v2 CHECK (kind IN ('patch', 'fix', 'meta'));

-- +goose Down
ALTER TABLE game_jobs DROP CONSTRAINT IF EXISTS game_jobs_kind_check_v2;
ALTER TABLE game_jobs ADD CONSTRAINT game_jobs_kind_check CHECK (kind IN ('patch', 'fix'));
ALTER TABLE games DROP COLUMN IF EXISTS cover_json;
ALTER TABLE games DROP COLUMN IF EXISTS description;
ALTER TABLE games DROP COLUMN IF EXISTS title;
