-- +goose Up
ALTER TYPE post_category RENAME TO post_category_old;

CREATE TYPE post_category AS ENUM(
    'PENDING',
    'INFRASTRUCTURE',
    'ECONOMY',
    'HEALTHCARE',
    'EDUCATION',
    'ENVIRONMENT',
    'SAFETY',
    'OTHER'
);


ALTER TABLE posts ALTER COLUMN category TYPE post_category USING category::text::post_category;

DROP TYPE post_category_old;

-- +goose Down
-- 1. Reassign new categories back to an old fallback ('GENERAL') so the downgrade doesn't crash
UPDATE posts SET category = 'GENERAL' WHERE category::text IN ('INFRASTRUCTURE', 'ENVIRONMENT', 'SAFETY', 'OTHER');

-- 2. Rename the current ENUM
ALTER TYPE post_category RENAME TO post_category_new;

-- 3. Recreate the original ENUM exactly as it was
CREATE TYPE post_category AS ENUM (
    'PENDING',
    'URBAN_PLANNING',
    'TRANSPORT',
    'ECONOMY',
    'HEALTHCARE',
    'EDUCATION',
    'GENERAL'
);

-- 4. Transition the column back
ALTER TABLE posts ALTER COLUMN category TYPE post_category USING category::text::post_category;

-- 5. Drop the newer ENUM
DROP TYPE post_category_new;
