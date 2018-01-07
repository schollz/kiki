package database

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
)

func Set(bucket, key string, value interface{}) (err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.Set(bucket, key, value)
}

func Get(bucket, key string, value interface{}) (err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.Get(bucket, key, value)
}

func AddEnvelope(e letter.Envelope) (err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.addEnvelope(e)
}

// GetEnvelopeFromID returns a single envelope from its ID
func GetEnvelopeFromID(id string) (e letter.Envelope, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	var es []letter.Envelope
	es, err = db.getAllFromPreparedQuery("SELECT * FROM letters WHERE id = ?", id)
	if err != nil {
		err = errors.Wrap(err, "GetEnvelopeFromID("+id+")")
	} else {
		e = es[0]
	}
	return
}

// GetAllEnvelopes returns all envelopes determined by whether they are opened
func GetAllEnvelopes(opened ...bool) (e []letter.Envelope, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	if len(opened) > 0 {
		if opened[0] {
			return db.getAllFromQuery("SELECT * FROM letters WHERE opened == 1")
		} else {
			return db.getAllFromQuery("SELECT * FROM letters WHERE opened == 0")
		}
	} else {
		return db.getAllFromQuery("SELECT * FROM letters")
	}
}

func (d *Database) getKeys() (s []keypair.KeyPair, err error) {
	// prepare statement
	query := "SELECT keypair FROM keypairs ORDER BY time DESC"
	log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "GetKeys")
		return
	}
	defer rows.Close()

	// parse rows
	s = make([]keypair.KeyPair, 100000)
	sI := 0
	// loop through rows
	for rows.Next() {
		var mKeyPair string
		err = rows.Scan(&mKeyPair)
		if err != nil {
			err = errors.Wrap(err, "getRows")
			return
		}

		kp, err := keypair.FromPublic(mKeyPair)
		if err != nil {
			log.Warn(err)
			continue
		}
		s[sI] = kp
		sI++
	}
	s = s[:sI]
	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getRows")
	}
	return
}
