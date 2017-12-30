package database

import (
	"log"
	"sync"

	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/envelope"
)

type Database struct {
	file string
	db   *storm.DB
	sync.RWMutex
}

// Setup a new database
func Setup(file string) (d *Database) {
	d = new(Database)
	d.Lock()
	defer d.Unlock()
	d.file = file
	return
}

func (d *Database) AddEnvelope(e *envelope.Envelope) (err error) {
	d.Lock()
	defer d.Unlock()
	d.db, err = storm.Open(d.file)
	defer d.db.Close()
	err = d.db.Save(e)
	return
}

func (d *Database) AddUnsealedEnvelope(e *envelope.UnsealedEnvelope) (err error) {
	d.Lock()
	defer d.Unlock()
	d.db, err = storm.Open(d.file)
	defer d.db.Close()
	err = d.db.Save(e)
	return
}

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

func (d *Database) Open() (err error) {
	d.Lock()
	d.db, err = storm.Open(d.file)
	return
}

func (d *Database) Close() (err error) {
	err = d.db.Close()
	d.Unlock()
	return
}

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

func (d *Database) Catalog() (catalog []string, err error) {
	err = d.Open()
	if err != nil {
		return
	}
	defer d.Close()
	count, err := d.db.Count(new(envelope.Envelope))
	if err != nil {
		err = errors.Wrap(err, "problem counting")
		return
	}
	log.Println(count)
	catalog = make([]string, count)
	i := 0
	query := d.db.Select()
	err = query.Each(new(envelope.Envelope), func(record interface{}) error {
		u := record.(*envelope.Envelope)
		catalog[i] = u.ID
		i++
		return nil
	})
	log.Println(catalog)
	if err != nil {
		err = errors.Wrap(err, "problem querying")
	}
	return
}
