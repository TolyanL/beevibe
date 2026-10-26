-- +goose Up
ALTER TABLE games ADD COLUMN IF NOT EXISTS slug TEXT;

UPDATE games
SET slug = 'game-' || substr(replace(id::text, '-', ''), 1, 8)
WHERE slug IS NULL OR slug = '';

ALTER TABLE games ALTER COLUMN slug SET NOT NULL;

DROP INDEX IF EXISTS games_slug_unique;
CREATE UNIQUE INDEX games_slug_unique ON games (slug);

-- +goose Down
DROP INDEX IF EXISTS games_slug_unique;
ALTER TABLE games DROP COLUMN IF EXISTS slug;
