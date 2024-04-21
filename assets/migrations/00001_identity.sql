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
    role    user_role NOT NULL DEFAULT 'maintainer',
    user_id INTEGER REFERENCES users (id) ON DELETE CASCADE,
    team_id INTEGER REFERENCES teams (id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, team_id)
);
-- CreateTeam function; ensure only 'admin' can create teams
CREATE OR REPLACE FUNCTION create_team(
    team_name TEXT,
    team_email TEXT,
    current_user_id TEXT
)
    RETURNS TABLE (
                      team_id TEXT,
                      name TEXT,
                      email TEXT,
                      updated_at TIMESTAMP
                  )
AS
$$
DECLARE
    new_team_id TEXT;
    new_name TEXT;
    new_email TEXT;
    new_updated_at TIMESTAMP;
BEGIN
    -- Check if the current user has the 'admin' role
    IF EXISTS (
        SELECT 1
        FROM user_team_mapping utm
                 JOIN users u ON utm.user_id = u.id
        WHERE u.user_id = current_user_id
          AND utm.role = 'admin'
    ) THEN
        -- Insert the new team
        INSERT INTO teams (name, email)
        VALUES (team_name, team_email)
        RETURNING teams.team_id, teams.name, teams.email, teams.updated_at
            INTO new_team_id, new_name, new_email, new_updated_at;

        -- Return the inserted team
        RETURN QUERY SELECT new_team_id, new_name, new_email, new_updated_at;
    ELSE
        RAISE EXCEPTION 'Only admins can create teams';
    END IF;
END;
$$
    LANGUAGE plpgsql VOLATILE;;
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
DROP FUNCTION create_team(team_name TEXT, team_email TEXT, current_user_id TEXT);
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
