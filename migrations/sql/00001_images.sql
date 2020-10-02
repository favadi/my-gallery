-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE images
(
    id      SERIAL PRIMARY KEY,
    name    TEXT NOT NULL DEFAULT '' UNIQUE,
    format  TEXT NOT NULL DEFAULT '',
    created TIMESTAMPTZ   DEFAULT now()
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE images;
