-- name: GetUserPermissions :many
SELECT permissions.code
FROM permissions
    INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
    INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.id = $1;