package database

import (
	"fmt"
	"sync"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/envelope"
	"github.com/schollz/kiki/src/logging"
)

// Database is a thread-safe wrapper to the asdine/storm database that provides functionality as a keystore and to set and get envelopes.
type Database struct {
	file string
	db   *storm.DB
	sync.RWMutex
}

// Setup a new database, used once on a global instance.
func Setup(file string) (d *Database) {
	logging.Log.Info("starting setup")
	d = new(Database)
	d.Lock()
	defer d.Unlock()
	d.file = file
	return
}

// Open locks and opens database. Must be called everytime you need to interact with the database.
func (d *Database) Open() (err error) {
	d.Lock()
	d.db, err = storm.Open(d.file)
	if err != nil {
		err = errors.Wrap(err, "problem opening db: '"+d.file+"'")
	}
	return
}

// Close closes the databases and unlocks the structure.
func (d *Database) Close() (err error) {
	err = d.db.Close()
	d.Unlock()
	return
}

// AddEnvelope adds an envelope to the database.
func (d *Database) AddEnvelope(e *envelope.Envelope) (err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()
	err = d.db.Save(e)
	return
}

// AddUnsealedEnvelope adds an unsealed envelope to the database.
func (d *Database) AddUnsealedEnvelope(e *envelope.UnsealedEnvelope) (err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()
	err = d.db.Save(e)
	return
}

// GetEnvelope allows you to get an enevelope by the id
func (d *Database) GetEnvelope(id string) (e *envelope.Envelope, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()
	e = new(envelope.Envelope)
	err = d.db.One("ID", id, e)
	return
}

// GetUnsealedEnvelope allows you to get an enevelope by the id
func (d *Database) GetUnsealedEnvelope(id string) (e *envelope.UnsealedEnvelope, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()
	e = new(envelope.UnsealedEnvelope)
	err = d.db.One("ID", id, e)
	return
}

// GetEnvelopes gets all of the sealed envelopes
func (d *Database) GetEnvelopes() (e []*envelope.Envelope, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()

	// get count
	query := d.db.Select().OrderBy("Timestamp")
	count, err := d.db.Count(new(envelope.Envelope))
	if err != nil {
		err = errors.Wrap(err, "problem counting")
		return
	}
	// pre make array
	e = make([]*envelope.Envelope, count)

	// collect all of them
	i := 0
	query = d.db.Select().OrderBy("Timestamp")
	err = query.Each(new(envelope.Envelope), func(record interface{}) error {
		u := record.(*envelope.Envelope)
		e[i] = u
		i++
		return nil
	})
	if err != nil {
		err = errors.Wrap(err, "problem querying")
	}
	return
}

// GetUnsealedEnvelopes gets all of the unsealed envelopes
func (d *Database) GetUnsealedEnvelopes() (e []*envelope.UnsealedEnvelope, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()

	// get count
	query := d.db.Select().OrderBy("Timestamp")
	count, err := d.db.Count(new(envelope.UnsealedEnvelope))
	if err != nil {
		err = errors.Wrap(err, "problem counting")
		return
	}
	// pre make array
	e = make([]*envelope.UnsealedEnvelope, count)

	// collect all of them
	i := 0
	query = d.db.Select().OrderBy("Timestamp")
	err = query.Each(new(envelope.UnsealedEnvelope), func(record interface{}) error {
		u := record.(*envelope.UnsealedEnvelope)
		e[i] = u
		i++
		return nil
	})
	if err != nil {
		err = errors.Wrap(err, "problem querying")
	}
	return
}

// EnvelopeCatalog returns a list of the current IDs of all the envelopes.
func (d *Database) EnvelopeCatalog() (catalog map[string]struct{}, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()

	// loop over each element
	catalog = make(map[string]struct{})
	query := d.db.Select()
	err = query.Each(new(envelope.Envelope), func(record interface{}) error {
		u := record.(*envelope.Envelope)
		catalog[u.ID] = struct{}{}
		return nil
	})
	if err != nil {
		err = errors.Wrap(err, "problem querying")
	}
	return
}

// UnsealedEnvelopeCatalog returns a list of the current IDs of all the unsealed envelopes.
func (d *Database) UnsealedEnvelopeCatalog() (catalog map[string]struct{}, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()

	// loop over each element
	catalog = make(map[string]struct{})
	query := d.db.Select()
	err = query.Each(new(envelope.UnsealedEnvelope), func(record interface{}) error {
		u := record.(*envelope.UnsealedEnvelope)
		catalog[u.ID] = struct{}{}
		return nil
	})
	if err != nil {
		err = errors.Wrap(err, "problem querying")
	}
	return
}

// Set stores a value in the keystore
func (d *Database) Set(bucket string, key interface{}, value interface{}) (err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()
	err = d.db.Set(bucket, key, value)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("problem setting '%s' in '%s'", key, bucket))
	}
	return
}

// Get returns a value from the key store
func (d *Database) Get(bucket string, key interface{}, to interface{}) (err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()

	err = d.db.Get(bucket, key, &to)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("problem getting '%s' from '%s'", key, bucket))
	}
	return
}

// Delete will remove a value from the keystore
func (d *Database) Delete(bucket string, key interface{}) (err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()

	err = d.db.Delete(bucket, key)
	return
}
