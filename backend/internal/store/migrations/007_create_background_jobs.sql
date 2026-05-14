-- +goose Up
-- +goose StatementBegin

CREATE TABLE background_jobs (
    id UUID PRIMARY KEY,
    job_type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT valid_status CHECK (status IN('PENDING','PROCESSING','COMPLETED','FAILED'))
);

CREATE INDEX idx_background_jobs_status on background_jobs(status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS background_jobs;
-- +goose StatementEnd
