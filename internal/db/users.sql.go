// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: users.sql

package db

import (
	"context"
)

const activateUser = `-- name: ActivateUser :one
UPDATE users
SET 
    activated = true,
    version = version + 1
WHERE id = $1 AND version = $2
RETURNING id, name, email, hashed_password, activated, version, created_at
`

type ActivateUserParams struct {
	UserID  int64 `json:"user_id"`
	Version int32 `json:"-"`
}

func (q *Queries) ActivateUser(ctx context.Context, arg ActivateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, activateUser, arg.UserID, arg.Version)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.HashedPassword,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (name, email, hashed_password, activated)
VALUES ($1, $2, $3, $4)
RETURNING id, name, email, hashed_password, activated, version, created_at
`

type CreateUserParams struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	HashedPassword []byte `json:"-"`
	Activated      bool   `json:"activated"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.Name,
		arg.Email,
		arg.HashedPassword,
		arg.Activated,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.HashedPassword,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, hashed_password, activated, version, created_at FROM users WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.HashedPassword,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}

const getUserByToken = `-- name: GetUserByToken :one
SELECT users.id, users.name, users.email, users.hashed_password, users.activated, users.version, users.created_at
FROM users
    INNER JOIN tokens ON users.id = tokens.user_id
WHERE tokens.hash = $1
    AND tokens.scope = $2
    AND tokens.expires_at > now()
`

type GetUserByTokenParams struct {
	Hash  []byte `json:"hash"`
	Scope string `json:"scope"`
}

func (q *Queries) GetUserByToken(ctx context.Context, arg GetUserByTokenParams) (User, error) {
	row := q.db.QueryRow(ctx, getUserByToken, arg.Hash, arg.Scope)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.HashedPassword,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
	)
	return i, err
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users
SET 
    hashed_password = $1,
    version = version + 1
WHERE id = $2 AND version = $3
RETURNING id, name, email, hashed_password, activated, version, created_at
`

type UpdateUserPasswordParams struct {
	HashedPassword []byte `json:"-"`
	UserID         int64  `json:"user_id"`
	Version        int32  `json:"-"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.Exec(ctx, updateUserPassword, arg.HashedPassword, arg.UserID, arg.Version)
	return err
}
