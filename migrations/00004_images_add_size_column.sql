-- +goose Up
-- SQL in this section is executed when the migration is applied.

ALTER TABLE images
    ADD COLUMN size INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

ALTER TABLE images
    DROP COLUMN size;
