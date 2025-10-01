-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, url, name, user_id)
VALUES(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
)
RETURNING *;

-- name: GetAllFeeds :many
SELECT * FROM feeds;
