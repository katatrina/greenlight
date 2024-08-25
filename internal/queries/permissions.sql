-- name: GetUserPermissions :many
SELECT permissions.code
FROM permissions
    INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
    INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.id = $1;

-- name: AddPermissionsForUser :exec
INSERT INTO users_permissions (user_id, permission_id)
SELECT sqlc.arg(user_id)::bigint, permissions.id FROM permissions
    WHERE permissions.code = ANY(sqlc.arg(permission_codes)::text[]);