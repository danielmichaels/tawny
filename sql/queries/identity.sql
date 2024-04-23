-- Create admin user (for initial setup only)
-- name: DoesAdminExist :one
SELECT EXISTS (SELECT 1
               FROM users u
                        JOIN team_user tu ON u.uuid = tu.user_id
               WHERE u.name = 'admin'
                 AND tu.role = 'admin') AS admin_exists;
-- Update a user role
-- name: UpdateUserRole :exec
UPDATE team_user
SET role = $1
WHERE user_id = (SELECT tokenable_id
                 FROM personal_access_tokens
                 WHERE token = $2);

-- Create a new user and a team for them
-- name: CreateUserWithNewTeam :one
WITH new_team AS (
    INSERT INTO teams (name, personal_team)
        VALUES (COALESCE($1 || '_team', 'default_team'), true)
        RETURNING uuid, id),
     new_user AS (
         INSERT INTO users (name, email, password, current_team_id)
             VALUES ($2, $3, $4, (SELECT id FROM new_team))
             RETURNING uuid),
     new_user_team AS (
         INSERT INTO team_user (team_id, user_id, role)
             SELECT new_team.uuid, new_user.uuid, 'maintainer'
             FROM new_team,
                  new_user
             RETURNING user_id, team_id),
     new_token AS (
         INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token)
             SELECT 'user', new_user.uuid, 'default', ('key_' || generate_uid(12))
             FROM new_user
             RETURNING token)
SELECT new_user.uuid AS user_id, new_team.uuid AS team_id, new_token.token AS personal_access_token
FROM new_user,
     new_team,
     new_token;

-- Get users in the same team mapping as the logged-in user when provided another user's ID
-- name: GetUserByID :one
SELECT uuid, name, email, created_at, updated_at
FROM users
WHERE id = (SELECT tokenable_id
            FROM personal_access_tokens
            WHERE token = $1);

-- Count all users the authorized user can see; used in pagination
-- name: CountUsers :one
SELECT count(*) OVER ()
FROM users u
         JOIN team_user tu ON u.uuid = tu.user_id
WHERE tu.team_id IN (SELECT team_id
                     FROM team_user
                     WHERE user_id = (SELECT tokenable_id
                                      FROM personal_access_tokens
                                      WHERE token = $1));

-- List all users associated with the authorized user and get the total count
-- name: ListUsers :many
SELECT u.id, u.name, u.email, u.created_at, u.updated_at, tu.role
FROM users u
         JOIN team_user tu ON u.id = tu.user_id
WHERE tu.team_id IN (SELECT team_id
                     FROM team_user
                     WHERE user_id = (SELECT tokenable_id
                                      FROM personal_access_tokens
                                      WHERE token = $1))
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;

-- Create a team; leverages 'create_team' function which when supplied
-- name, email, and user_id will either create a team or error on permissions
-- name: CreateTeam :one
SELECT *
FROM create_team($1, $2, $3);

-- Retrieve user with team info (used in API-KEY auth)
-- name: RetrieveUserWithTeamInfoByAPIKEY :one
SELECT u.uuid       AS user_uuid,
       u.name       AS username,
       u.email      AS user_email,
       u.created_at AS user_created_at,
       u.updated_at AS user_updated_at,
       t.uuid       AS team_uuid,
       t.name       AS team_name,
       t.created_at AS team_created_at,
       t.updated_at AS team_updated_at
FROM users u
         JOIN team_user tu ON u.id = tu.user_id
         JOIN teams t ON tu.team_id = t.id
WHERE u.id = (SELECT tokenable_id
              FROM personal_access_tokens
              WHERE token = $1);
