-- name: GetAccount :one
SELECT * FROM account
WHERE id = ? LIMIT 1;

-- name: ListAccount :many
SELECT * FROM account;

-- name: CreateAccount :one
INSERT INTO account (
    id, name, updatedAt, createdAt
) VALUES (
             ?, ?, ?, ?
         )
    RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = ?;

-- Users
-- name: GetUser :one
SELECT * FROM user
WHERE id = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM user
WHERE email = ? LIMIT 1;

-- name: ListUser :many
SELECT * FROM user;

-- name: CreateUser :one
INSERT INTO user (
   id, email, name, account, updatedAt, createdAt
) VALUES (
           ?,  ?, ?, ?, ?, ?
         )
    RETURNING *;

-- name: DeleteUser :exec
DELETE FROM user
WHERE id = ?;