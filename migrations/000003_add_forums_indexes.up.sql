-- Filename: migrations/000003_add_forums_indexes.up.sql
CREATE INDEX IF NOT EXISTS forums_topic_idx ON forums USING GIN(to_tsvector('simple', topic));