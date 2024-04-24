// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package store

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRole string

const (
	UserRoleAdmin      UserRole = "admin"
	UserRoleMaintainer UserRole = "maintainer"
)

func (e *UserRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserRole(s)
	case string:
		*e = UserRole(s)
	default:
		return fmt.Errorf("unsupported scan type for UserRole: %T", src)
	}
	return nil
}

type NullUserRole struct {
	UserRole UserRole `json:"user_role"`
	Valid    bool     `json:"valid"` // Valid is true if UserRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserRole) Scan(value interface{}) error {
	if value == nil {
		ns.UserRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserRole), nil
}

type PersonalAccessTokens struct {
	ID            int64              `json:"id"`
	TokenableType string             `json:"tokenable_type"`
	TokenableID   string             `json:"tokenable_id"`
	Name          string             `json:"name"`
	Token         string             `json:"token"`
	Abilities     pgtype.Text        `json:"abilities"`
	LastUsedAt    pgtype.Timestamptz `json:"last_used_at"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

type TeamUser struct {
	ID        int32              `json:"id"`
	TeamID    pgtype.Text        `json:"team_id"`
	UserID    pgtype.Text        `json:"user_id"`
	Role      UserRole           `json:"role"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	UpdatedAt pgtype.Timestamptz `json:"updated_at"`
}

type Teams struct {
	ID           int32              `json:"id"`
	Uuid         string             `json:"uuid"`
	PersonalTeam pgtype.Bool        `json:"personal_team"`
	Name         string             `json:"name"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}

type Users struct {
	ID               int32              `json:"id"`
	Uuid             string             `json:"uuid"`
	Name             pgtype.Text        `json:"name"`
	Email            pgtype.Text        `json:"email"`
	EmailVerifiedAt  pgtype.Timestamptz `json:"email_verified_at"`
	Password         pgtype.Text        `json:"password"`
	RememberToken    pgtype.Text        `json:"remember_token"`
	CurrentTeamID    pgtype.Int4        `json:"current_team_id"`
	ProfilePhotoPath pgtype.Text        `json:"profile_photo_path"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	UpdatedAt        pgtype.Timestamptz `json:"updated_at"`
}
