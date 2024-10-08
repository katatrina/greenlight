-- name: CreateToken :one
INSERT INTO tokens (user_id, hash, scope, expires_at)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteUserTokens :exec
DELETE FROM tokens
WHERE user_id = $1 AND scope = $2;