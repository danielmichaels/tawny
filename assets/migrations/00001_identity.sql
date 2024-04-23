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
-- Teams Table
CREATE TABLE teams
(
    id            SERIAL PRIMARY KEY,
    uuid          TEXT UNIQUE                 NOT NULL DEFAULT ('team_' || generate_uid(12)),
    personal_team BOOLEAN,
    name          TEXT                        NOT NULL DEFAULT '',
    created_at    TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- Users Table
CREATE TABLE users
(
    id                 SERIAL PRIMARY KEY,
    uuid               TEXT UNIQUE                 NOT NULL DEFAULT ('user_' || generate_uid(12)),
    name               VARCHAR(255),
    email              VARCHAR(255) UNIQUE,
    email_verified_at  TIMESTAMP(0) WITH TIME ZONE NULL,
    password           VARCHAR(255),
    remember_token     VARCHAR(100),
    current_team_id    INTEGER REFERENCES teams (id),
    profile_photo_path VARCHAR(2048),
    created_at         TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);
-- User-Team Mapping Table
CREATE TABLE team_user
(
    id         SERIAL PRIMARY KEY,
    team_id    TEXT REFERENCES teams (uuid) ON DELETE CASCADE,
    user_id    TEXT REFERENCES users (uuid) ON DELETE CASCADE,
    role       user_role                   NOT NULL DEFAULT 'maintainer',
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (team_id, user_id)
);
CREATE TABLE personal_access_tokens
(
    id             BIGSERIAL PRIMARY KEY,
    tokenable_type VARCHAR(255)                NOT NULL,
    tokenable_id   TEXT                        NOT NULL,
    name           VARCHAR(255)                NOT NULL,
    token          VARCHAR(64)                 NOT NULL,
    abilities      TEXT                        NULL,
    last_used_at   TIMESTAMP(0) WITH TIME ZONE NULL,
    created_at     TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT personal_access_tokens_token_unique UNIQUE (token),
    CONSTRAINT fk_tokenable_id FOREIGN KEY (tokenable_id) REFERENCES users (uuid) ON DELETE CASCADE
);
CREATE INDEX personal_access_tokens_tokenable_id_index
    ON personal_access_tokens (tokenable_id);
-- Add index to team_id column in team_user table
CREATE INDEX team_user_team_id_idx ON team_user (team_id);
-- Add index to user_id column in team_user table
CREATE INDEX team_user_user_id_idx ON team_user (user_id);
-- CreateTeam function; ensure only 'admin' can create teams
CREATE OR REPLACE FUNCTION create_team(
    team_name TEXT,
    team_email TEXT,
    current_user_id TEXT
)
    RETURNS TABLE
            (
                team_id    TEXT,
                name       TEXT,
                email      TEXT,
                updated_at TIMESTAMP
            )
AS
$$
DECLARE
    new_team_id    TEXT;
    new_name       TEXT;
    new_email      TEXT;
    new_updated_at TIMESTAMP;
BEGIN
    -- Check if the current user has the 'admin' role
    IF EXISTS (SELECT 1
               FROM team_user tm
                        JOIN users u ON tm.user_id = u.id
               WHERE u.id = current_user_id
                 AND tm.role = 'admin') THEN
        -- Insert the new team
        INSERT INTO teams (name, email)
        VALUES (team_name, team_email)
        RETURNING teams.id, teams.name, teams.updated_at
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
CREATE TRIGGER trigger_updated_at_team_user
    BEFORE UPDATE
    ON team_user
    FOR EACH ROW
EXECUTE FUNCTION updated_at_trigger();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_user;
DROP TABLE IF EXISTS personal_access_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
DROP FUNCTION create_team(team_name TEXT, team_email TEXT, current_user_id TEXT);
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd
