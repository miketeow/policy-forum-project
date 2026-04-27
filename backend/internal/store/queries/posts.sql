-- name: CreatePost :one
INSERT INTO posts (id, user_id, title, content, category, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetPostByID :one
SELECT posts.id, posts.title, posts.content, posts.category, posts.created_at, posts.updated_at, posts.score,
    users.id AS author_id, users.name AS author_name,
    COALESCE(pv.vote,0)::smallint AS user_vote
FROM posts
JOIN users ON posts.user_id = users.id
LEFT JOIN post_votes pv ON pv.post_id = posts.id AND pv.user_id = sqlc.narg('current_user_id')::uuid
WHERE posts.id = $1 LIMIT 1;

-- name: ListPostsByNewest :many
SELECT posts.id, posts.title, posts.category, posts.created_at, posts.updated_at, posts.score,
    users.name AS author_name, users.id AS author_id,
    COALESCE(pv.vote,0)::smallint AS user_vote
FROM posts
JOIN users ON posts.user_id = users.id
LEFT JOIN post_votes pv ON pv.post_id = posts.id AND pv.user_id = sqlc.narg('current_user_id')::uuid
WHERE (sqlc.narg('category')::post_category IS NULL OR posts.category = sqlc.narg('category'))
AND (sqlc.narg('cursor')::timestamp IS NULL OR posts.created_at < sqlc.narg('cursor'))
ORDER BY posts.created_at DESC
LIMIT $1;

-- name: ListPostsByOldest :many
SELECT posts.id, posts.title, posts.category, posts.created_at, posts.updated_at, posts.score,
    users.name AS author_name, users.id AS author_id,
    COALESCE(pv.vote,0)::smallint AS user_vote
FROM posts
JOIN users ON posts.user_id = users.id
LEFT JOIN post_votes pv ON pv.post_id = posts.id AND pv.user_id = sqlc.narg('current_user_id')::uuid
WHERE (sqlc.narg('category')::post_category IS NULL OR posts.category = sqlc.narg('category'))
AND (sqlc.narg('cursor')::timestamp IS NULL OR posts.created_at > sqlc.narg('cursor'))
ORDER BY posts.created_at ASC
LIMIT $1;

-- name: UpdatePost :one
UPDATE posts
SET title = $3, content = $4, category = $5, updated_at = $6
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts
WHERE id = $1 AND user_id = $2;

-- name: GetPostVote :one
SELECT vote from post_votes WHERE post_id = $1 AND user_id = $2;

-- name: SetPostVote :exec
-- This is "Upsert". If the row exists, it overwrites the vote. If not, it inserts it
INSERT INTO post_votes (post_id, user_id, vote)
VALUES ($1, $2, $3)
ON CONFLICT (post_id, user_id) DO UPDATE
SET vote = EXCLUDED.vote;

-- name: RemovePostVote :exec
DELETE FROM post_votes WHERE post_id = $1 AND user_id = $2;

-- name: UpdatePostScore :exec
UPDATE posts SET score = score + $2 WHERE id = $1;
