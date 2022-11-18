-- Filename: migrations/000002_create_forums_table.up.sql

CREATE TABLE IF NOT EXISTS forums (
    id bigserial PRIMARY KEY,
    username text NOT NULL, 
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    topic text NOT NULL,
    discussion text NOT NULL,
    comments text[],
    total_comments int NOT NULL DEFAULT 0,
    version integer NOT NULL DEFAULT 1
);