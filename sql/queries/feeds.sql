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



