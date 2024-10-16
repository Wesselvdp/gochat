-- name: GetUser :one
SELECT * FROM users
WHERE email = ? LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY moment;

-- name: CreateUser :one
INSERT INTO users (
 email, company
) VALUES (
  ?, ?
)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE email = ?;