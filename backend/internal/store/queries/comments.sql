-- name: CreateComments :one
INSERT INTO comments (id, post_id, user_id, parent_id,content, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetCommentsByPostID :many
SELECT comments.id, comments.parent_id, comments.content, comments.created_at, comments.updated_at, users.id AS author_id, users.name AS author_name
FROM comments
JOIN users ON comments.user_id = users.id
WHERE comments.post_id = $1
ORDER BY comments.created_at DESC;
