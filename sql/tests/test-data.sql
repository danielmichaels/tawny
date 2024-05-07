-- ADMIN Team
WITH new_team AS (
    INSERT INTO teams (uuid, name, personal_team)
        VALUES ('team_0000000', COALESCE('admin' || '_team', 'default_team'), true)
        RETURNING uuid, id),
     new_user AS (
         INSERT INTO users (uuid, name, email, password, current_team_id)
             VALUES ('user_0000000', 'admin', 'admin@tawny.internal', crypt('password', gen_salt('bf')),
                     (SELECT id FROM new_team))
             RETURNING uuid),
     new_user_team AS (
         INSERT INTO team_user (team_id, user_id, role)
             SELECT new_team.uuid, new_user.uuid, 'admin'
             FROM new_team,
                  new_user
             RETURNING user_id, team_id),
     new_token AS (
         INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token)
             SELECT 'user', new_user.uuid, 'default', 'key_00000000000000000000'
             FROM new_user
             RETURNING token)
SELECT new_user.uuid AS user_id, new_team.uuid AS team_id, new_token.token AS personal_access_token
FROM new_user,
     new_team,
     new_token;
;
-- User_1 Team
WITH new_team AS (
    INSERT INTO teams (uuid, name, personal_team)
        VALUES ('team_0000001', COALESCE('user' || '_team', 'default_team'), true)
        RETURNING uuid, id),
     new_user AS (
         INSERT INTO users (uuid, name, email, password, current_team_id)
             VALUES ('user_0000001', 'user', 'user@tawny.internal', crypt('password', gen_salt('bf')),
                     (SELECT id FROM new_team))
             RETURNING uuid),
     new_user_team AS (
         INSERT INTO team_user (team_id, user_id, role)
             SELECT new_team.uuid, new_user.uuid, 'maintainer'
             FROM new_team,
                  new_user
             RETURNING user_id, team_id),
     new_token AS (
         INSERT INTO personal_access_tokens (tokenable_type, tokenable_id, name, token)
             SELECT 'user', new_user.uuid, 'default', 'key_00000000000000000001'
             FROM new_user
             RETURNING token)
SELECT new_user.uuid AS user_id, new_team.uuid AS team_id, new_token.token AS personal_access_token
FROM new_user,
     new_team,
     new_token;
;
-- todo create non personal teams
