-- alter feeds table to add last_fetched_at column that's a timestamptz and can be null
-- +goose Up
ALTER TABLE feeds
ADD COLUMN IF NOT EXISTS last_fetched_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE feeds
DROP COLUMN IF EXISTS last_fetched_at;

