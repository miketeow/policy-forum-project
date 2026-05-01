-- name: GlobalSearch :many
SELECT 'post'::text AS result_type, p.id AS unique_id, p.id as url_id, p.title, p.content, p.category, u.name AS author_name, p.created_at
FROM posts p
JOIN users u ON p.user_id = u.id
WHERE to_tsvector('english', p.title || ' ' || p.content || ' ' || p.category) @@ plainto_tsquery('english', @search_query::text)

UNION ALL

SELECT 'comment'::text AS result_type, c.id as unique_id, c.post_id AS url_id, p.title AS title, c.content, p.category AS category, u.name AS author_name, c.created_at
FROM comments c
JOIN users u ON c.user_id = u.id
JOIN posts p ON c.post_id = p.id
WHERE to_tsvector('english', c.content || ' ' || p.category) @@ plainto_tsquery('english', @search_query::text)

ORDER BY created_at DESC
Limit 20;
