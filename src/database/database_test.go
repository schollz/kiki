package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/schollz/kiki/src/envelope"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/person"
	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	p, err := person.New()
	assert.Nil(t, err)
	l, err := letter.New("post", "hello world", p.Keys.Public)
	assert.Nil(t, err)
	e, err := envelope.New(l, p, []*person.Person{p})
	assert.Nil(t, err)

	os.Remove("test.db")
	d := Setup("test.db")
	err = d.AddEnvelope(e)
	assert.Nil(t, err)
	err = d.AddEnvelope(e)
	assert.Nil(t, err)
	catalog, err := d.Catalog()
	assert.Nil(t, err)
	assert.Equal(t, 1, len(catalog))

	e2, err := d.GetEnvelope(e.ID)
	assert.Nil(t, err)
	fmt.Println(e2.ID)
	assert.Equal(t, *e.Sender, *e2.Sender)
	assert.Equal(t, e.Recipients, e2.Recipients)
}

func TestReading(t *testing.T) {
	p, err := person.New()
	assert.Nil(t, err)
	l, err := letter.New("post", "hello world", p.Keys.Public)
	assert.Nil(t, err)
	e, err := envelope.New(l, p, []*person.Person{p})
	assert.Nil(t, err)
	u, err := e.Unseal([]*person.Person{p})
	assert.Nil(t, err)

	os.Remove("test.db")
	d := Setup("test.db")
	err = d.AddUnsealedEnvelope(u)
	assert.Nil(t, err)

	l, err = letter.New("post", "hello world, again", p.Keys.Public)
	assert.Nil(t, err)
	e, err = envelope.New(l, p, []*person.Person{p})
	u, err = e.Unseal([]*person.Person{p})
	assert.Nil(t, err)
	err = d.AddUnsealedEnvelope(u)
	assert.Nil(t, err)

	es, err := d.GetUnsealedEnvelopes()
	assert.Nil(t, err)
	assert.Equal(t, "hello world", es[0].Letter.Content.Data)
	assert.Equal(t, "hello world, again", es[1].Letter.Content.Data)
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
