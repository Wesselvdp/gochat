-- name: GetAccount :one
SELECT * FROM account
WHERE id = ? LIMIT 1;

-- name: GetAccountByDomain :one
SELECT a.id
FROM account a
         JOIN account_domain ad ON a.id = ad.account
WHERE ad.domain = ?;

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

-- name: GetUserByExternalID :one
SELECT * FROM user
WHERE externalId = ? LIMIT 1;

-- name: ListUser :many
SELECT * FROM user;

-- name: CreateUser :one
INSERT INTO user (
   id, email, externalId, name, account, updatedAt, createdAt
) VALUES (
           ?,  ?, ?, ?, ?, ?, ?
         )
    RETURNING *;

-- name: DeleteUser :exec
DELETE FROM user
WHERE id = ?;


-- EVENTS

-- name: CreateEvent :one
INSERT INTO event (
    user, event, metadata
) VALUES (
     ?, ?, ?
)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM event
WHERE id = ? LIMIT 1;


-- FILES
-- name: CreateFile :one
INSERT INTO file (
    id, name, owner
) VALUES (
 ?, ?, ?
 )
RETURNING *;

-- name: GetFile :one
SELECT * FROM file
WHERE id = ? LIMIT 1;

