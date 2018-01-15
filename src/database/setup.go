package database

import (
	"database/sql"
	"github.com/mattn/go-sqlite3"
)

func init() {
	sql.Register("sqlite3_with_extensions",
		&sqlite3.SQLiteDriver{
			Extensions: []string{
				"json1",
			},
		})
}
