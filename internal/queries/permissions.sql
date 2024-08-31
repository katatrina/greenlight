-- name: GetUserPermissions :many
SELECT permissions.code
FROM users
INNER JOIN users_permissions ON users.id = users_permissions.user_id
INNER JOIN permissions ON users_permissions.permission_id = permissions.id
WHERE users.id = $1;

-- name: AddPermissionsForUser :exec
INSERT INTO users_permissions (user_id, permission_id)
SELECT sqlc.arg(user_id)::bigint, permissions.id FROM permissions
    WHERE permissions.code = ANY(sqlc.arg(permission_codes)::text[]);