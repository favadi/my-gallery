-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE TABLE users
(
    id            SERIAL PRIMARY KEY,
    username      TEXT        NOT NULL DEFAULT '' UNIQUE,
    password_hash TEXT        NOT NULL DEFAULT '',
    full_name     TEXT        NOT NULL DEFAULT '',
    created       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated       TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO users(username, password_hash, full_name)
VALUES ('demo1', '$2y$12$pRKXfflFLHUAok7iihRHBuk1x34fkwdFE2qc06LnZTCtPOVE8XdaO', 'Demo User 1'),
       ('demo2', '$2y$12$bbyyU.P164NaVF/Q.oHrPO7s9FtQkRSFa/zJtzLOjzgMVTpeZ8HQK', 'Demo User 2'),
       ('demo3', '$2y$12$h.98mawu4p.iWkVCrGE9guNVZbUnf9crp3WNLYNTHJiPSbs3.Hopi', 'Demo User 3');

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE users;
