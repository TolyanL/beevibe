-- +goose Up
ALTER TABLE game_jobs DROP CONSTRAINT IF EXISTS game_jobs_kind_check;

-- +goose Down
SELECT 1;
