package models

import (
	"database/sql"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
)

type MyNullString struct {
	sql.NullString
}

func (s MyNullString) MarshalJSON() ([]byte, error) {
	if s.Valid {
		return json.Marshal(s.String)
	}
	return []byte(`null`), nil
}

type Claims struct {
	UserID  string `json:"sub"`
	LocalID int64  `json:"localID"`
	jwt.RegisteredClaims
}
