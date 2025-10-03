-- name: GetUserByEmail :one
SELECT id,
       email,
       password,
       tag
FROM users
WHERE email = $1;



