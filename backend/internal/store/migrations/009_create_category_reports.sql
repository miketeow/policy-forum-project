-- +goose Up

-- create an append only table to track sentiment change over time
-- instead of overwriting a single row per category

CREATE TABLE category_reports(
    id UUID PRIMARY KEY,
    category post_category NOT NULL,
    report JSONB NOT NULL,
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_category_reports_latest ON category_reports(category, generated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS category_reports;
