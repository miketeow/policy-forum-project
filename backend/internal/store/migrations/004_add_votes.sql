-- +goose Up
ALTER TABLE posts ADD COLUMN score INT NOT NULL DEFAULT 0;
ALTER TABLE comments ADD COLUMN score INT NOT NULL DEFAULT 0;

CREATE TABLE post_votes (
    post_id UUID NOT NULL REFERENCES posts(id) on DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) on DELETE CASCADE,
    vote SMALLINT NOT NULL CHECK (vote IN (-1,1)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id,user_id)
);

CREATE TABLE comment_votes (
    comment_id UUID NOT NULL REFERENCES comments(id) on DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) on DELETE CASCADE,
    vote SMALLINT NOT NULL CHECK (vote IN (-1,1)),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (comment_id,user_id)
);

-- +goose Down
DROP TABLE IF EXISTS comment_votes;
DROP TABLE IF EXISTS post_votes;
ALTER TABLE comments DROP COLUMN score;
ALTER TABLE posts DROP COLUMN score;
