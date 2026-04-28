-- name: CreateComments :one
INSERT INTO comments (id, post_id, user_id, parent_id,content, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListCommentsByNewest :many
SELECT comments.id, comments.parent_id, comments.content, comments.created_at, comments.updated_at,
    comments.score,
    users.id AS author_id, users.name AS author_name,
    COALESCE(cv.vote, 0)::smallint AS user_vote,
    (SELECT COUNT(*) FROM comments AS replies WHERE replies.parent_id = comments.id) AS reply_count
FROM comments
JOIN users ON comments.user_id = users.id
LEFT JOIN comment_votes cv ON cv.comment_id = comments.id AND cv.user_id = sqlc.narg('current_user_id')::uuid
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
SELECT comments.id, comments.parent_id, comments.content, comments.created_at, comments.updated_at,
    comments.score,
    users.id AS author_id, users.name AS author_name,
    COALESCE(cv.vote, 0)::smallint AS user_vote,
    (SELECT COUNT(*) FROM comments AS replies WHERE replies.parent_id = comments.id) AS reply_count
FROM comments
JOIN users ON comments.user_id = users.id
LEFT JOIN comment_votes cv ON cv.comment_id = comments.id AND cv.user_id = sqlc.narg('current_user_id')::uuid
WHERE comments.post_id = $1
AND(
(sqlc.narg('parent_id')::uuid IS NULL AND comments.parent_id IS NULL)
OR
(comments.parent_id = sqlc.narg('parent_id'))
)
AND (sqlc.narg('cursor')::timestamp IS NULL OR comments.created_at > sqlc.narg('cursor'))
ORDER BY comments.created_at ASC
LIMIT $2;

-- name: ListCommentsByPopular :many
SELECT comments.id, comments.parent_id, comments.content, comments.created_at, comments.updated_at,
    comments.score,
    users.id AS author_id, users.name AS author_name,
    COALESCE(cv.vote, 0)::smallint AS user_vote,
    (SELECT COUNT(*) FROM comments AS replies WHERE replies.parent_id = comments.id) AS reply_count
FROM comments
JOIN users ON comments.user_id = users.id
LEFT JOIN comment_votes cv ON cv.comment_id = comments.id AND cv.user_id = sqlc.narg('current_user_id')::uuid
WHERE comments.post_id = $1
AND (comments.parent_id = sqlc.narg('parent_id')::uuid OR (sqlc.narg('parent_id')::uuid IS NULL AND comments.parent_id IS NULL))
ORDER BY comments.score DESC, comments.created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateComment :one
UPDATE comments
SET content = $3, updated_at = $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1 AND user_id = $2;

-- name: GetCommentVote :one
SELECT vote FROM comment_votes WHERE comment_id = $1 AND user_id = $2;

-- name: SetCommentVote :exec
INSERT INTO comment_votes(comment_id,user_id,vote)
VALUES($1,$2,$3)
ON CONFLICT (comment_id,user_id) DO UPDATE
SET vote = EXCLUDED.vote;

-- name: RemoveCommentVote :exec
DELETE FROM comment_votes WHERE comment_id = $1 AND user_id = $2;

-- name: UpdateCommentScore :exec
UPDATE comments SET score = score + $2 WHERE id = $1;
