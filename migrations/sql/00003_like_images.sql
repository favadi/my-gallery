-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE likes
(
    id       SERIAL PRIMARY KEY,
    user_id  INTEGER     NOT NULL REFERENCES users (id),
    image_id INTEGER     NOT NULL REFERENCES images (id),
    created  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX like_uni ON likes (user_id, image_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE likes;
