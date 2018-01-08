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
			return db.getAllFromQuery("SELECT * FROM letters WHERE opened == 1 ORDER BY time DESC")
		} else {
			return db.getAllFromQuery("SELECT * FROM letters WHERE opened == 0 ORDER BY time DESC")
		}
	} else {
		return db.getAllFromQuery("SELECT * FROM letters ORDER BY time DESC")
	}
}

// GetKeys will return all the keys
func GetKeys() (s []keypair.KeyPair, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.getKeys()
}

// GetKeysFromSender will return all the keys from a certain sender
func GetKeysFromSender(sender string) (s []keypair.KeyPair, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.getKeys(sender)
}

// GetName will return the assigned name for the public key of a sender
func GetName(publicKey string) (name string, err error) {
	db, err := Open()
	if err != nil {
		return
	}
	defer db.Close()
	return db.getName(publicKey)
}
