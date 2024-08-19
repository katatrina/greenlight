-- name: CreateToken :exec
INSERT INTO tokens (user_id, hash, scope, expired_at)
VALUES ($1, $2, $3, $4);

-- name: DeleteUserTokens :exec
DELETE FROM tokens
WHERE user_id = $1 AND scope = $2;