-- Filename: migrations/000005_add_permissions_table.up.sql

CREATE TABLE IF NOT EXISTS permissions (
    id bigserial PRIMARY KEY,
    code text NOT NULL
);

-- create a linking table that link users to permissions
-- this is an example of many to many relationship
CREATE TABLE IF NOT EXISTS users_permissions (
    user_id BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions (id) ON DELETE CASCADE,
    PRIMARY KEY(user_id, permission_id)
);

INSERT INTO permissions (code) 
VALUES 
('forums:read'), ('forums:write');
