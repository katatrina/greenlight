-- name: CreateUser :one
INSERT INTO users (name, email, hashed_password, activated)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET 
    name = $2,
    email = $3,
    hashed_password = $4,
    activated = $5,
    version = version + 1
WHERE id = $1 AND version = $6
RETURNING *;
