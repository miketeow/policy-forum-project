-- +goose Up
-- Add a nullable parent_id to support nested replies
ALTER TABLE comments
ADD COLUMN parent_id UUID REFERENCES comments(id) ON DELETE CASCADE;

-- create an index to quickly find all replies to a specific comments
CREATE INDEX idx_comments_parent_id ON comments(parent_id);

-- +goose Down
ALTER TABLE comments DROP COLUMN parent_id;
