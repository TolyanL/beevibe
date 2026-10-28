-- +goose Up
ALTER TABLE games ADD COLUMN IF NOT EXISTS resume_stage TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE games DROP COLUMN IF EXISTS resume_stage;
