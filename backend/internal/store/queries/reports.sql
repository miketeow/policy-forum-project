-- name: SaveCategoryReport :exec
INSERT INTO category_reports (id, category, report, generated_at)
VALUES ($1, $2, $3, $4);

-- name: GetLatestCategoryReport :one
SELECT id, category, report, generated_at
FROM category_reports
WHERE category = $1
ORDER BY generated_at DESC
LIMIT 1;
