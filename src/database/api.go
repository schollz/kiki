package database

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/schollz/kiki/src/letter"
)

func AddEnvelope(e letter.Envelope) (err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.addEnvelope(e)
}

func Set(bucket, key string, value interface{}) (err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.Set(bucket, key, value)
}
