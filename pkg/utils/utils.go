package utils

import (
	"database/sql"
	"time"
)

func GetTime() sql.NullString {
	return sql.NullString{
		String: time.Now().UTC().Format(time.RFC3339),
		Valid:  true,
	}
}
