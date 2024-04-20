// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: identity.sql

package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTeam = `-- name: CreateTeam :one
INSERT INTO teams (name, email)
VALUES ($1, $2)
RETURNING team_id, name, email, updated_at
`

type CreateTeamParams struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateTeamRow struct {
	TeamID    string             `json:"team_id"`
	Name      string             `json:"name"`
	Email     string             `json:"email"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

// Create a team
func (q *Queries) CreateTeam(ctx context.Context, arg CreateTeamParams) (CreateTeamRow, error) {
	row := q.db.QueryRow(ctx, createTeam, arg.Name, arg.Email)
	var i CreateTeamRow
	err := row.Scan(
		&i.TeamID,
		&i.Name,
		&i.Email,
		&i.UpdatedAt,
	)
	return i, err
}

const createUserWithNewTeam = `-- name: CreateUserWithNewTeam :one
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
     new_team
`

type CreateUserWithNewTeamParams struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserWithNewTeamRow struct {
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
}

// Create a new user and a team for them
func (q *Queries) CreateUserWithNewTeam(ctx context.Context, arg CreateUserWithNewTeamParams) (CreateUserWithNewTeamRow, error) {
	row := q.db.QueryRow(ctx, createUserWithNewTeam, arg.Name, arg.Email)
	var i CreateUserWithNewTeamRow
	err := row.Scan(&i.UserID, &i.TeamID)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT user_id,
       username,
       email,
       verified,
       created_at,
       updated_at
FROM users
WHERE user_id = $1
`

type GetUserByIDRow struct {
	UserID    string             `json:"user_id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Verified  bool               `json:"verified"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

// Get a user
func (q *Queries) GetUserByID(ctx context.Context, userID string) (GetUserByIDRow, error) {
	row := q.db.QueryRow(ctx, getUserByID, userID)
	var i GetUserByIDRow
	err := row.Scan(
		&i.UserID,
		&i.Username,
		&i.Email,
		&i.Verified,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const listUsers = `-- name: ListUsers :many
SELECT u.id, u.username, u.email, u.verified
FROM users u
         JOIN user_team_mapping utm ON u.id = utm.user_id
WHERE utm.team_id IN (SELECT utm_inner.team_id
                      FROM user_team_mapping utm_inner
                      WHERE utm_inner.user_id = $1)
`

type ListUsersRow struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

// List all users associated to authorized user
func (q *Queries) ListUsers(ctx context.Context, userID int32) ([]ListUsersRow, error) {
	rows, err := q.db.Query(ctx, listUsers, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ListUsersRow{}
	for rows.Next() {
		var i ListUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.Email,
			&i.Verified,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}