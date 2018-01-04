package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	flock "github.com/theckman/go-flock"
)

var (
	log          = logging.Log
	DatabaseFile = "kiki.sqlite3.db"
)

type Database struct {
	name     string
	db       *sql.DB
	fileLock *flock.Flock
}

func Setup(locationToDatabase string) {
	DatabaseFile = locationToDatabase
}

// Open will open the database for transactions by first aquiring a filelock.
func Open(readOnly ...bool) (d *Database, err error) {
	d = new(Database)
	name := "kiki"

	// convert the name to base64 for file writing
	d.name = DatabaseFile

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

// Get will retrieve the value associated with a key.
func (d *Database) Get(bucket, key string, v interface{}) (err error) {
	stmt, err := d.db.Prepare("select value from keystore where bucket_key = ?")
	if err != nil {
		return errors.Wrap(err, "problem preparing SQL")
	}
	defer stmt.Close()
	var result string
	err = stmt.QueryRow(bucket + "/" + key).Scan(&result)
	if err != nil {
		return errors.Wrap(err, "problem getting key")
	}

	err = json.Unmarshal([]byte(result), &v)
	if err != nil {
		return
	}
	return
}

// Set will set a value in the database, when using it like a keystore.
func (d *Database) Set(bucket, key string, value interface{}) (err error) {
	var b []byte
	b, err = json.Marshal(value)
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "Set")
	}
	stmt, err := tx.Prepare("insert or replace into keystore(bucket_key,value) values (?, ?)")
	if err != nil {
		return errors.Wrap(err, "Set")
	}
	defer stmt.Close()

	_, err = stmt.Exec(bucket+"/"+key, string(b))
	if err != nil {
		return errors.Wrap(err, "Set")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Set")
	}

	return
}

// addEnvelope will add or replace an envelope
func (d *Database) addEnvelope(e letter.Envelope) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return
	}
	var opened int
	// marshaled things
	var mSender, mSealedRecipients, mTo, mChannels string
	if e.Opened {
		opened = 1
	} else {
		opened = 0
	}
	var b []byte
	b, err = json.Marshal(e.Sender)
	if err != nil {
		return errors.Wrap(err, "problem marshaling Sender")
	}
	mSender = string(b)

	b, err = json.Marshal(e.SealedRecipients)
	if err != nil {
		return errors.Wrap(err, "problem marshaling SealedRecipients")
	}
	mSealedRecipients = string(b)

	b, err = json.Marshal(e.Letter.To)
	if err != nil {
		return errors.Wrap(err, "problem marshaling To")
	}
	mTo = string(b)

	b, err = json.Marshal(e.Letter.Channels)
	if err != nil {
		return errors.Wrap(err, "problem marshaling Channels")
	}
	mChannels = string(b)

	stmt, err := tx.Prepare("insert or replace into letters(id,time,sender,signature,sealed_recipients,sealed_letter,opened,letter_purpose,letter_to,letter_content,letter_replaces,letter_channels,letter_replyto) values(?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.ID, e.Timestamp, mSender, e.Signature, mSealedRecipients, e.SealedLetter, opened, e.Letter.Purpose, mTo, e.Letter.Content, e.Letter.Replaces, mChannels, e.Letter.ReplyTo)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

// GetEnvelopeFromID returns a single envelope from its ID
func (d *Database) GetEnvelopeFromID(id string) (e letter.Envelope, err error) {
	var es []letter.Envelope
	es, err = d.getAllFromPreparedQuery("SELECT * FROM letters WHERE id = ?", id)
	if err != nil {
		err = errors.Wrap(err, "GetEnvelopeFromID("+id+")")
	} else {
		e = es[0]
	}
	return
}

// GetAllEnvelopes returns all envelopes determined by whether they are opened
func (d *Database) GetAllEnvelopes(opened bool) (e []letter.Envelope, err error) {
	if opened {
		return d.getAllFromQuery("SELECT * FROM letters WHERE opened == 1")
	} else {
		return d.getAllFromQuery("SELECT * FROM letters WHERE opened == 0")
	}
}

func (d *Database) getAllFromQuery(query string) (s []letter.Envelope, err error) {
	log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getAllFromQuery")
		return
	}
	defer rows.Close()

	// parse rows
	s, err = d.getRows(rows)
	if err != nil {
		err = errors.Wrap(err, query)
	}
	return
}

// getAllFromPreparedQuery
func (d *Database) getAllFromPreparedQuery(query string, args ...interface{}) (s []letter.Envelope, err error) {
	// prepare statement
	stmt, err := d.db.Prepare(query)
	if err != nil {
		err = errors.Wrap(err, query)
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		err = errors.Wrap(err, query)
		return
	}
	defer rows.Close()
	s, err = d.getRows(rows)
	if err != nil {
		err = errors.Wrap(err, query)
	}
	return
}

func (d *Database) getRows(rows *sql.Rows) (s []letter.Envelope, err error) {
	s = make([]letter.Envelope, 100000)
	sI := 0
	// loop through rows
	for rows.Next() {
		var e letter.Envelope
		e.Letter = letter.Letter{}
		var opened int
		// marshaled things
		var mSender, mSealedRecipients, mTo, mChannels string
		err = rows.Scan(&e.ID, &e.Timestamp, &mSender, &e.Signature, &mSealedRecipients, &e.SealedLetter, &opened, &e.Letter.Purpose, &mTo, &e.Letter.Content, &e.Letter.Replaces, &mChannels, &e.Letter.ReplyTo)
		json.Unmarshal([]byte(mSender), &e.Sender)
		json.Unmarshal([]byte(mSealedRecipients), &e.SealedRecipients)
		json.Unmarshal([]byte(mTo), &e.Letter.To)
		json.Unmarshal([]byte(mChannels), &e.Letter.Channels)

		e.Opened = opened == 1
		if err != nil {
			err = errors.Wrap(err, "getRows")
			return
		}

		s[sI] = e
		sI++
	}
	s = s[:sI]
	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getRows")
	}
	return
}
