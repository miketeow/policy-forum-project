-- +goose Up
-- +goose StatementBegin
ALTER TABLE post_votes
ADD CONSTRAINT unique_post_user_vote UNIQUE (post_id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE post_votes
DROP CONSTRAINT unique_post_user_vote;
-- +goose StatementEnd
