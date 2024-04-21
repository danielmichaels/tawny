-- Create admin user (for initial setup only)
-- name: DoesAdminExist :one
SELECT EXISTS (SELECT 1
               FROM users u
                        JOIN user_team_mapping utm ON u.id = utm.user_id
               WHERE u.username = 'admin'
                 AND utm.role = 'admin') AS admin_exists;

-- Update a user role
-- name: UpdateUserRole :exec
UPDATE user_team_mapping
SET role = $1
WHERE user_id = (SELECT id
                 FROM users u
                 WHERE u.user_id = $2);

-- Create a new user and a team for them
-- name: CreateUserWithNewTeam :one
WITH new_team AS (
    INSERT INTO teams (name, email)
        VALUES ($1, $2)
        RETURNING id,team_id),
     new_user AS (
         INSERT INTO users (username, email, password, verified)
             VALUES ($1, $2, $3, false)
             RETURNING id,user_id),
     new_user_team as (
         INSERT INTO user_team_mapping (user_id, team_id)
             SELECT new_user.id, new_team.id
             FROM new_user,
                  new_team
             RETURNING user_id, team_id)
SELECT new_user.user_id, new_team.team_id
FROM new_user,
     new_team;

-- Get a user
-- name: GetUserByID :one
SELECT user_id,
       username,
       email,
       verified,
       created_at,
       updated_at
FROM users
WHERE user_id = $1;

-- List all users associated to authorized user
-- name: ListUsers :many
SELECT u.id, u.username, u.email, u.verified, u.created_at, u.updated_at, utm.role
FROM users u
         JOIN user_team_mapping utm ON u.id = utm.user_id
WHERE utm.team_id IN (SELECT utm_inner.team_id
                      FROM user_team_mapping utm_inner
                               JOIN users u_inner ON utm_inner.user_id = u_inner.id
                      WHERE u_inner.user_id = $1);

-- Create a team
-- name: CreateTeam :one
INSERT INTO teams (name, email)
VALUES ($1, $2)
RETURNING team_id, name, email, updated_at;

-- Retrieve user with team info (used in API-KEY auth)
-- name: RetrieveUserWithTeamInfoByAPIKEY :one
SELECT u.id         AS user_pk,
       u.user_id,
       u.username,
       u.email      AS user_email,
       u.verified   AS user_verified,
       u.created_at AS user_created_at,
       u.updated_at AS user_updated_at,
       t.id         AS team_pk,
       t.team_id    AS team_id,
       t.name       AS team_name,
       t.email      AS team_email,
       t.created_at AS team_created_at,
       t.updated_at AS team_updated_at
FROM users u
         JOIN
     user_team_mapping ut ON u.id = ut.user_id
         JOIN
     teams t ON ut.team_id = t.id
WHERE u.api_key = $1;