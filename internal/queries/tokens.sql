-- name: CreateToken :exec
INSERT INTO tokens (hash, user_id, expiry, scope)
VALUES ($1, $2, $3, $4);