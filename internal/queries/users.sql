-- name: CreateUser :one
INSERT INTO users (name, email, hashed_password, activated)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ActivateUser :one
UPDATE users
SET 
    activated = true,
    version = version + 1
WHERE id = sqlc.arg(user_id) AND version = sqlc.arg(version)
RETURNING *;

-- name: GetUserByToken :one
SELECT users.*
FROM users
    INNER JOIN tokens ON users.id = tokens.user_id
WHERE tokens.hash = $1
    AND tokens.scope = $2
    AND tokens.expired_at > now();