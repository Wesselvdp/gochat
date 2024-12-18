-- name: GetAccount :one
SELECT
    a.*,
    IFNULL(GROUP_CONCAT(ad.domain), '') AS domains
FROM
    account a
        LEFT JOIN
    account_domain ad
    ON
        ad.account = a.id
WHERE
    a.id = ?
GROUP BY
    a.id;


-- name: GetAccountByDomain :one
SELECT a.id
FROM account a
         JOIN account_domain ad ON a.id = ad.account
WHERE ad.domain = ?;

-- name: DeleteAccountDomain :exec
DELETE FROM account_domain
WHERE domain = sqlc.arg(domain);

-- name: ListAccount :many
SELECT * FROM account;

-- name: CreateAccount :one
INSERT INTO account (
    id, name
) VALUES (
             ?, ?
         )
    RETURNING *;

-- name: CreateAccountDomain :one
INSERT INTO account_domain (
    account, domain
) VALUES (
             ?, ?
         )
    RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = ?;

-- Users
-- name: GetUser :one
SELECT
    u.*,
    a.id AS account_id,
    a.name AS account_name
FROM user u
         LEFT JOIN account a ON u.account = a.id
WHERE u.id = ?
LIMIT 1;

-- name: UpdateUserAccount :exec
UPDATE user
SET account = sqlc.arg(accountID),
    updatedAt = datetime('now')
WHERE id = sqlc.arg(userID);

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
   id, email, externalId, name, account
) VALUES (
           ?,  ?, ?, ?, ?
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

