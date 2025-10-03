-- name: DeleteRefreshTokensByUserID :exec
DELETE FROM user_tokens
WHERE user_id = $1
  AND category = 'REFRESH';

-- name: CreateRefreshToken :one
INSERT INTO user_tokens (user_id, category, hash, expiry)
VALUES ($1, $2, $3, $4)
    RETURNING id, user_id, category, hash, expiry;