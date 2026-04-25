-- name: CreateComments :one
INSERT INTO comments (id, post_id, user_id, parent_id,content, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListCommentsByNewest :many
SELECT comments.id, comments.parent_id, comments.content, comments.created_at, comments.updated_at, users.id AS author_id, users.name AS author_name,
    (SELECT COUNT(*) FROM comments AS replies WHERE replies.parent_id = comments.id) AS reply_count
FROM comments
JOIN users ON comments.user_id = users.id
WHERE comments.post_id = $1
AND(
(sqlc.narg('parent_id')::uuid IS NULL AND comments.parent_id IS NULL)
OR
(comments.parent_id = sqlc.narg('parent_id'))
)
AND (sqlc.narg('cursor')::timestamp IS NULL OR comments.created_at < sqlc.narg('cursor'))
ORDER BY comments.created_at DESC
LIMIT $2;

-- name: ListCommentsByOldest :many
SELECT comments.id, comments.parent_id, comments.content, comments.created_at, comments.updated_at, users.id AS author_id, users.name AS author_name,
    (SELECT COUNT(*) FROM comments AS replies WHERE replies.parent_id = comments.id) AS reply_count
FROM comments
JOIN users ON comments.user_id = users.id
WHERE comments.post_id = $1
AND(
(sqlc.narg('parent_id')::uuid IS NULL AND comments.parent_id IS NULL)
OR
(comments.parent_id = sqlc.narg('parent_id'))
)
AND (sqlc.narg('cursor')::timestamp IS NULL OR comments.created_at > sqlc.narg('cursor'))
ORDER BY comments.created_at ASC
LIMIT $2;
