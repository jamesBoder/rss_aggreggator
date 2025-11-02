-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at, feed_id, title, url, description, published_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at, feed_id, title, url, description, published_at;

-- name: GetPostsForUser :many
SELECT posts.id, posts.created_at, posts.updated_at, posts.feed_id, posts.title, posts.url, posts.description, posts.published_at
FROM posts
INNER JOIN feeds ON posts.feed_id = feeds.id
WHERE feeds.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2 OFFSET $3;