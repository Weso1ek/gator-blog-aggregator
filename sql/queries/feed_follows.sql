-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
    )
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    feed.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
LEFT JOIN feed ON feed.id = inserted_feed_follow.feed_id
LEFT JOIN users ON users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
SELECT feed.name AS feed_name, users.name AS user_name
FROM feed_follows
LEFT JOIN feed ON feed.id = feed_follows.feed_id
LEFT JOIN users ON users.id = feed_follows.user_id
WHERE feed_follows.user_id = $1;