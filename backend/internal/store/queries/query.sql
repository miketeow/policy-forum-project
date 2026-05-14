-- name: EnqueueJob :exec
INSERT INTO background_jobs (id, job_type, payload, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: DequeueJob :one
-- Finds the oldest PENDING job, locks it, mark it PROCESSING, and hands it to the Go Worker
UPDATE background_jobs
SET status = 'PROCESSING', updated_at = $1
WHERE id = (
    SELECT id
    FROM background_jobs
    WHERE status = 'PENDING'
    ORDER BY created_at ASC
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
RETURNING id, job_type, payload;

-- name: CompleteJob :exec
UPDATE background_jobs
SET status = 'COMPLETED', updated_at = $2
WHERE id = $1;

-- name: FailJob :exec
UPDATE background_jobs
SET status = 'FAILED', error_message = $2, updated_at = $3
WHERE id = $1;

-- name: GetPostForWorker :one
-- query for background jobs, no joins
SELECT id, title, content
FROM posts
WHERE id = $1 LIMIT 1;

-- name: UpdatePostSummary :exec
-- save the finished summary back to the posts
UPDATE posts
SET summary = $2, updated_at = $3
WHERE id = $1;

-- name: CheckSummaryJobStatus :one
-- used by the SSE endpoint to check if frontend need update UI
SELECT status, created_at
FROM background_jobs
WHERE job_type = 'EXEC_SUMMARY'
AND payload->>'post_id' = $1::text
ORDER BY created_at DESC
LIMIT 1;
