package database

import (
	"database/sql"
	"fmt"
	"os"
	"path"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/logging"
	flock "github.com/theckman/go-flock"
)

var DataFolder = "."

var (
	log = logging.Log
)

type Database struct {
	name     string
	db       *sql.DB
	fileLock *flock.Flock
}

// Open will open the database for transactions by first aquiring a filelock.
func Open(readOnly ...bool) (d *Database, err error) {
	d = new(Database)
	name := "kiki"

	// convert the name to base64 for file writing
	d.name = path.Join(DataFolder, name+".sqlite3.db")

	// if read-only, make sure the database exists
	if _, err = os.Stat(d.name); err != nil && len(readOnly) > 0 && readOnly[0] {
		err = errors.New(fmt.Sprintf("group '%s' does not exist", name))
		return
	}

	// obtain a lock on the database
	d.fileLock = flock.NewFlock(d.name + ".lock")
	for {
		locked, err := d.fileLock.TryLock()
		if err == nil && locked {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// check if it is a new database
	newDatabase := false
	if _, err := os.Stat(d.name); os.IsNotExist(err) {
		newDatabase = true
	}

	// open sqlite3 database
	d.db, err = sql.Open("sqlite3", d.name)
	if err != nil {
		return
	}

	// create new database tables if needed
	if newDatabase {
		err = d.MakeTables()
		if err != nil {
			return
		}
	}

	return
}

// Close will close the database connection and remove the filelock.
func (d *Database) Close() (err error) {
	// close filelock
	err = d.fileLock.Unlock()
	if err != nil {
		log.Error(err)
	} else {
		os.Remove(d.name + ".lock")
		log.Info(err)
	}
	// close database
	err2 := d.db.Close()
	if err2 != nil {
		err = err2
		log.Error(err)
	} else {
		log.Info("closed database")
	}
	return
}

// MakeTables creates two tables, a `keystore` table:
//
// 	BUCKET_KEY (TEXT)	VALUE (TEXT)
//
// and also a `letters`:
func (d *Database) MakeTables() (err error) {
	sqlStmt := `create table keystore (bucket_key text not null primary key, value text);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}
	sqlStmt = `create index keystore_idx on keystore(bucket_key);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}
	sqlStmt = `create table letters (id text not null primary key, time TIMESTAMP, sender text, signature text, sealed_recipients text, sealed_letter text, opened integer, letter_purpose text, letter_to text, letter_content text, letter_replaces text, letter_channels text, letter_replyto text, unique(id));`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}
	return
}
