// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: queries.sql

package schema

import (
	"context"
	"database/sql"
)

const createAccount = `-- name: CreateAccount :one
INSERT INTO account (
    id, name, updatedAt, createdAt
) VALUES (
             ?, ?, ?, ?
         )
    RETURNING id, name, createdat, updatedat
`

type CreateAccountParams struct {
	ID        string
	Name      string
	Updatedat string
	Createdat string
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount,
		arg.ID,
		arg.Name,
		arg.Updatedat,
		arg.Createdat,
	)
	var i Account
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
	)
	return i, err
}

const createAccountDomain = `-- name: CreateAccountDomain :one
INSERT INTO account_domain (
    account, domain
) VALUES (
             ?, ?
         )
    RETURNING account, domain
`

type CreateAccountDomainParams struct {
	Account string
	Domain  string
}

func (q *Queries) CreateAccountDomain(ctx context.Context, arg CreateAccountDomainParams) (AccountDomain, error) {
	row := q.db.QueryRowContext(ctx, createAccountDomain, arg.Account, arg.Domain)
	var i AccountDomain
	err := row.Scan(&i.Account, &i.Domain)
	return i, err
}

const createEvent = `-- name: CreateEvent :one

INSERT INTO event (
    user, event, metadata
) VALUES (
     ?, ?, ?
)
RETURNING id, event, timestamp, metadata, user
`

type CreateEventParams struct {
	User     string
	Event    string
	Metadata interface{}
}

// EVENTS
func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (Event, error) {
	row := q.db.QueryRowContext(ctx, createEvent, arg.User, arg.Event, arg.Metadata)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.Event,
		&i.Timestamp,
		&i.Metadata,
		&i.User,
	)
	return i, err
}

const createFile = `-- name: CreateFile :one
INSERT INTO file (
    id, name, owner
) VALUES (
 ?, ?, ?
 )
RETURNING id, name, createdat, updatedat, owner
`

type CreateFileParams struct {
	ID    string
	Name  string
	Owner string
}

// FILES
func (q *Queries) CreateFile(ctx context.Context, arg CreateFileParams) (File, error) {
	row := q.db.QueryRowContext(ctx, createFile, arg.ID, arg.Name, arg.Owner)
	var i File
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Owner,
	)
	return i, err
}

const createUser = `-- name: CreateUser :one
INSERT INTO user (
   id, email, externalId, name, account, updatedAt, createdAt
) VALUES (
           ?,  ?, ?, ?, ?, ?, ?
         )
    RETURNING id, name, email, account, externalid, createdat, updatedat
`

type CreateUserParams struct {
	ID         string
	Email      string
	Externalid sql.NullString
	Name       sql.NullString
	Account    string
	Updatedat  string
	Createdat  string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.ID,
		arg.Email,
		arg.Externalid,
		arg.Name,
		arg.Account,
		arg.Updatedat,
		arg.Createdat,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Account,
		&i.Externalid,
		&i.Createdat,
		&i.Updatedat,
	)
	return i, err
}

const deleteAccount = `-- name: DeleteAccount :exec
DELETE FROM account
WHERE id = ?
`

func (q *Queries) DeleteAccount(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteAccount, id)
	return err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM user
WHERE id = ?
`

func (q *Queries) DeleteUser(ctx context.Context, id string) error {
	_, err := q.db.ExecContext(ctx, deleteUser, id)
	return err
}

const getAccount = `-- name: GetAccount :one
SELECT a.id, a.name, a.createdat, a.updatedat, ad.account, ad.domain
FROM account a
         LEFT JOIN account_domain ad ON ad.account = a.id
WHERE a.id = ? LIMIT 1
`

type GetAccountRow struct {
	ID        string
	Name      string
	Createdat string
	Updatedat string
	Account   sql.NullString
	Domain    sql.NullString
}

func (q *Queries) GetAccount(ctx context.Context, id string) (GetAccountRow, error) {
	row := q.db.QueryRowContext(ctx, getAccount, id)
	var i GetAccountRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Account,
		&i.Domain,
	)
	return i, err
}

const getAccountByDomain = `-- name: GetAccountByDomain :one
SELECT a.id
FROM account a
         JOIN account_domain ad ON a.id = ad.account
WHERE ad.domain = ?
`

func (q *Queries) GetAccountByDomain(ctx context.Context, domain string) (string, error) {
	row := q.db.QueryRowContext(ctx, getAccountByDomain, domain)
	var id string
	err := row.Scan(&id)
	return id, err
}

const getEvent = `-- name: GetEvent :one
SELECT id, event, timestamp, metadata, user FROM event
WHERE id = ? LIMIT 1
`

func (q *Queries) GetEvent(ctx context.Context, id int64) (Event, error) {
	row := q.db.QueryRowContext(ctx, getEvent, id)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.Event,
		&i.Timestamp,
		&i.Metadata,
		&i.User,
	)
	return i, err
}

const getFile = `-- name: GetFile :one
SELECT id, name, createdat, updatedat, owner FROM file
WHERE id = ? LIMIT 1
`

func (q *Queries) GetFile(ctx context.Context, id string) (File, error) {
	row := q.db.QueryRowContext(ctx, getFile, id)
	var i File
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Createdat,
		&i.Updatedat,
		&i.Owner,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT
    u.id, u.name, u.email, u.account, u.externalid, u.createdat, u.updatedat,
    a.id AS account_id,
    a.name AS account_name
FROM user u
         LEFT JOIN account a ON u.account = a.id
WHERE u.id = ?
LIMIT 1
`

type GetUserRow struct {
	ID          string
	Name        sql.NullString
	Email       string
	Account     string
	Externalid  sql.NullString
	Createdat   string
	Updatedat   string
	AccountID   sql.NullString
	AccountName sql.NullString
}

// Users
func (q *Queries) GetUser(ctx context.Context, id string) (GetUserRow, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i GetUserRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Account,
		&i.Externalid,
		&i.Createdat,
		&i.Updatedat,
		&i.AccountID,
		&i.AccountName,
	)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT id, name, email, account, externalid, createdat, updatedat FROM user
WHERE email = ? LIMIT 1
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Account,
		&i.Externalid,
		&i.Createdat,
		&i.Updatedat,
	)
	return i, err
}

const getUserByExternalID = `-- name: GetUserByExternalID :one
SELECT id, name, email, account, externalid, createdat, updatedat FROM user
WHERE externalId = ? LIMIT 1
`

func (q *Queries) GetUserByExternalID(ctx context.Context, externalid sql.NullString) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByExternalID, externalid)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.Account,
		&i.Externalid,
		&i.Createdat,
		&i.Updatedat,
	)
	return i, err
}

const listAccount = `-- name: ListAccount :many
SELECT id, name, createdat, updatedat FROM account
`

func (q *Queries) ListAccount(ctx context.Context) ([]Account, error) {
	rows, err := q.db.QueryContext(ctx, listAccount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Account
	for rows.Next() {
		var i Account
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Createdat,
			&i.Updatedat,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUser = `-- name: ListUser :many
SELECT id, name, email, account, externalid, createdat, updatedat FROM user
`

func (q *Queries) ListUser(ctx context.Context) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, listUser)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []User
	for rows.Next() {
		var i User
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Email,
			&i.Account,
			&i.Externalid,
			&i.Createdat,
			&i.Updatedat,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
