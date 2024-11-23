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

// Convert a string to sql.NullString
func StringToNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "", // Valid is true if the string is not empty
	}
}
