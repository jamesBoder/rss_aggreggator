-- name: GetAllFeeds :many
SELECT id, created_at, updated_at, name, url, user_id
FROM feeds;

-- name: GetFeedByID :one
SELECT id, created_at, updated_at, name, url, user_id
FROM feeds
WHERE id = $1;

-- name: GetFeedByURL :one
SELECT id, created_at, updated_at, name, url, user_id
FROM feeds
WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at
FROM feeds
ORDER BY last_fetched_at NULLS FIRST, updated_at ASC
LIMIT 1;



