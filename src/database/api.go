package database

import (
	"database/sql"
	"encoding/json"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/letter"
)

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

// AddEnvelope will add or replace an envelope
func (d *Database) AddEnvelope(e letter.Envelope) (err error) {
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
