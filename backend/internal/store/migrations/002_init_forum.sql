-- +goose Up

-- use strict ENUM to protect backend routing and frontend tabs
CREATE TYPE post_category AS ENUM (
    'PENDING',
    'URBAN_PLANNING',
    'TRANSPORT',
    'ECONOMY',
    'HEALTHCARE',
    'EDUCATION',
    'GENERAL'
);

-- create posts table
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category post_category NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- create comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- create indexes for foreign key ( crucial for performance when look up "All posts by user x")
CREATE INDEX idx_posts_user_id ON posts (user_id);
CREATE INDEX idx_comments_posts_id ON comments (post_id);
CREATE INDEX idx_comments_user_id ON comments (user_id);

-- +goose Down
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS posts;
DROP TYPE IF EXISTS post_category;
