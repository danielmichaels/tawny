-- Create a new user and a team for them
-- name: CreateUserWithNewTeam :one
WITH new_team AS (
    INSERT INTO teams (name, email)
        VALUES ($1, $2)
        RETURNING id,team_id),
     new_user AS (
         INSERT INTO users (username, email, verified)
             VALUES ($1, $2, false)
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
SELECT u.id, u.username, u.email, u.verified
FROM users u
         JOIN user_team_mapping utm ON u.id = utm.user_id
WHERE utm.team_id IN (SELECT utm_inner.team_id
                      FROM user_team_mapping utm_inner
                      WHERE utm_inner.user_id = $1);

-- Create a team
-- name: CreateTeam :one
INSERT INTO teams (name, email)
VALUES ($1, $2)
RETURNING team_id, name, email, updated_at;
