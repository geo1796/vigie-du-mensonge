-- name: GetRolesByUserID :many
SELECT r.id, r.name
FROM roles r
         JOIN user_roles ur ON ur.role_id = r.id
WHERE ur.user_id = $1;