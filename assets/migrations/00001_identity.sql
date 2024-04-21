-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- Helper Function: updated_at_trigger
CREATE OR REPLACE FUNCTION updated_at_trigger()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = current_timestamp(0);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- Helper Function: generate_uid
CREATE OR REPLACE FUNCTION generate_uid(size INT) RETURNS TEXT AS
$$
DECLARE
    characters TEXT  := 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    bytes      BYTEA := gen_random_bytes(size);
    l          INT   := length(characters);
    i          INT   := 0;
    output     TEXT  := '';
BEGIN
    WHILE i < size
        LOOP
            output := output || substr(characters, get_byte(bytes, i) % l + 1, 1);
            i := i + 1;
        END LOOP;
    RETURN output;
END;
$$ LANGUAGE plpgsql VOLATILE;
-- RBAC roles
CREATE TYPE user_role AS ENUM ('admin', 'maintainer');
-- Users Table
CREATE TABLE users
(
    id         SERIAL PRIMARY KEY,
    user_id    TEXT UNIQUE                 NOT NULL DEFAULT ('user_' || generate_uid(12)),
    username   VARCHAR(200)                NOT NULL,
    password   VARCHAR(255)                NOT NULL,
    email      VARCHAR(200)                NOT NULL UNIQUE,
    verified   BOOL                        NOT NULL DEFAULT false,
    api_key    VARCHAR(20)                 NOT NULL DEFAULT ('key_' || generate_uid(12)),
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_users_api_key ON users (api_key);
-- Teams Table
CREATE TABLE teams
(
    id         SERIAL PRIMARY KEY,
    team_id    TEXT UNIQUE                 NOT NULL DEFAULT ('team_' || generate_uid(12)),
    name       TEXT                        NOT NULL DEFAULT '',
    email      TEXT                        NOT NULL UNIQUE,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- User-Team Mapping Table
CREATE TABLE user_team_mapping
(
    role user_role NOT NULL DEFAULT 'maintainer',
    user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
    team_id INTEGER REFERENCES teams (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, team_id)
);
-- Triggers
CREATE TRIGGER trigger_updated_at_users
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION updated_at_trigger();
CREATE TRIGGER trigger_updated_at_teams
    BEFORE UPDATE
    ON teams
    FOR EACH ROW
EXECUTE FUNCTION updated_at_trigger();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_api_key;
DROP TABLE IF EXISTS user_team_mapping;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
