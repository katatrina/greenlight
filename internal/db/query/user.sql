-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, activated)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY id LIMIT $1
OFFSET $2;

-- name: ListTotalUsers :one
SELECT COUNT(*)
FROM users;