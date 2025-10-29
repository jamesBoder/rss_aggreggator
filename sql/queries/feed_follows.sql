-- CreateFeedFollow should insert a feed follow record then return all the fields from the feed follow as wll as the names of the linked user and feed.
-- name: CreateFeedFollow :one
WITH inserted AS (
  INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
  VALUES ($1, $2, $3, $4, $5)
  RETURNING *
)
SELECT
  inserted.*,
  feeds.name AS feed_name,
  users.name AS user_name
FROM inserted
JOIN feeds ON feeds.id = inserted.feed_id
JOIN users ON users.id = inserted.user_id;



-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.id,
    feed_follows.created_at,
    feed_follows.updated_at,
    feed_follows.user_id,
    feed_follows.feed_id,
    feeds.name AS feed_name,
    users.name AS user_name
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
INNER JOIN users ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;

-- name: DeleteFeedFollowByUserAndFeedID :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;