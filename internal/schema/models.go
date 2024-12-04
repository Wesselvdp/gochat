// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package schema

import (
	"database/sql"
)

type Account struct {
	ID        string
	Name      string
	Createdat string
	Updatedat string
}

type AccountDomain struct {
	Account string
	Domain  string
}

type Event struct {
	ID        int64
	Event     string
	Timestamp string
	Metadata  interface{}
	User      string
}

type File struct {
	ID        string
	Name      string
	Createdat string
	Updatedat string
	Owner     string
}

type User struct {
	ID         string
	Name       sql.NullString
	Email      string
	Account    string
	Externalid sql.NullString
	Createdat  string
	Updatedat  string
}
