// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: user.sql

package db

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, activated)
VALUES ($1, $2, $3, $4) RETURNING id, name, email, password_hash, activated, version, created_at, role
`

type CreateUserParams struct {
	Name         string
	Email        string
	PasswordHash []byte `json:"-"`
	Activated    bool
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.Name,
		arg.Email,
		arg.PasswordHash,
		arg.Activated,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PasswordHash,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
		&i.Role,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, activated, version, created_at, role
FROM users
WHERE email = $1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PasswordHash,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
		&i.Role,
	)
	return i, err
}

const getUserByID = `-- name: GetUserByID :one
SELECT id, name, email, password_hash, activated, version, created_at, role
FROM users
WHERE id = $1
`

func (q *Queries) GetUserByID(ctx context.Context, id int64) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.PasswordHash,
		&i.Activated,
		&i.Version,
		&i.CreatedAt,
		&i.Role,
	)
	return i, err
}

const listTotalUsers = `-- name: ListTotalUsers :one
SELECT COUNT(*)
FROM users
`

func (q *Queries) ListTotalUsers(ctx context.Context) (int64, error) {
	row := q.db.QueryRowContext(ctx, listTotalUsers)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const listUsers = `-- name: ListUsers :many
SELECT id, name, email, password_hash, activated, version, created_at, role
FROM users
ORDER BY id LIMIT $1
OFFSET $2
`

type ListUsersParams struct {
	Limit  int32
	Offset int32
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUsers, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.PasswordHash,
			&i.Activated,
			&i.Version,
			&i.CreatedAt,
			&i.Role,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
