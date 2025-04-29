-- name: CreateFeed :one
INSERT INTO feed (id, created_at, updated_at, name, url, user_id)
VALUES (
           $1,
           $2,
           $3,
           $4,
           $5,
           $6
       )
    RETURNING *;

-- name: GetFeeds :many
SELECT feed.name, feed.url, users.name FROM feed
LEFT JOIN users ON feed.user_id = users.id
ORDER BY feed.name ASC;

-- name: GetFeedsByUrl :one
SELECT * FROM feed
WHERE url = $1 LIMIT 1;

-- name: MarkFeedFetched :one
UPDATE feed
SET last_fetched_at = NOW(),
    updated_at = NOW()
WHERE id = $1
    RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT * FROM feed
ORDER BY last_fetched_at ASC NULLS FIRST
    LIMIT 1;