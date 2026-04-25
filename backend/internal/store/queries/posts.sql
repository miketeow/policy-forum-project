-- name: CreatePost :one
INSERT INTO posts (id, user_id, title, content, category, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPostByID :one
SELECT posts.id, posts.title, posts.content, posts.category, posts.created_at, posts.updated_at, users.id AS author_id, users.name AS author_name
FROM posts
JOIN users ON posts.user_id = users.id
WHERE posts.id = $1 LIMIT 1;

-- name: ListPostsByNewest :many
SELECT posts.id, posts.title, posts.category, posts.created_at, users.name AS author_name
FROM posts
JOIN users ON posts.user_id = users.id
WHERE (sqlc.narg('category')::post_category IS NULL OR posts.category = sqlc.narg('category'))
AND (sqlc.narg('cursor')::timestamp IS NULL OR posts.created_at < sqlc.narg('cursor'))
ORDER BY posts.created_at DESC
LIMIT $1;

-- name: ListPostsByOldest :many
SELECT posts.id, posts.title, posts.category, posts.created_at, users.name AS author_name
FROM posts
JOIN users ON posts.user_id = users.id
WHERE (sqlc.narg('category')::post_category IS NULL OR posts.category = sqlc.narg('category'))
AND (sqlc.narg('cursor')::timestamp IS NULL OR posts.created_at > sqlc.narg('cursor'))
ORDER BY posts.created_at ASC
LIMIT $1;
