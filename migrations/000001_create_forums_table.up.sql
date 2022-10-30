-- Filename: migrations/000001_create_forums_table.up.sql

CREATE TABLE IF NOT EXISTS forums (
    id bigserial PRIMARY KEY,
    users_id int DEFAULT 1,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    topic text NOT NULL,
    discussion text[] NOT NULL,
    version integer NOT NULL DEFAULT 1
);