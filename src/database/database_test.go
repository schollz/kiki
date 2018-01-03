package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/schollz/kiki/src/envelope"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/person"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	logging.Debug(false)
	p, err := person.New()
	regionkey, _ := person.New()
	assert.Nil(t, err)
	l := letter.NewText("hello, world")
	assert.Nil(t, err)
	e, err := envelope.New(l, p, []*person.Person{p}, regionkey)
	assert.Nil(t, err)

	os.Remove("test.db")
	d := Setup("test.db")
	err = d.AddEnvelope(e)
	assert.Nil(t, err)
	err = d.AddEnvelope(e)
	assert.Nil(t, err)
	catalog, err := d.EnvelopeCatalog()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(catalog))

	e2, err := d.GetEnvelope(e.ID)
	assert.Nil(t, err)
	fmt.Println(e2.ID)
	assert.Equal(t, *e.Sender, *e2.Sender)
	assert.Equal(t, e.Recipients, e2.Recipients)

	// test updating an envelope
	assert.Equal(t, false, e2.Opened)
	err = e2.Unseal([]*person.Person{p}, regionkey)
	assert.Nil(t, err)
	assert.Equal(t, "hello, world", e2.Letter.Text)
	assert.Equal(t, true, e2.Opened)
	err = d.AddEnvelope(e2)
	assert.Nil(t, err)
	e3, err := d.GetEnvelope(e.ID)
	assert.Equal(t, e2, e3)
	assert.NotEqual(t, e, e3)
}

func TestKeystore(t *testing.T) {
	os.Remove("test.db")
	d := Setup("test.db")
	type Something struct {
		Name string
	}
	s := new(Something)
	s.Name = "zack"
	err := d.Set("somethings", 1, &s)
	assert.Nil(t, err)

	s2 := new(Something)
	err = d.Get("somethings", 1, s2)
	assert.Nil(t, err)
	assert.Equal(t, *s, *s2)

	err = d.Delete("somethings", 1)
	assert.Nil(t, err)
	err = d.Get("somethings", 1, s2)
	assert.NotNil(t, err)
}
