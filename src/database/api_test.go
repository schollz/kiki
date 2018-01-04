package database

import (
	"os"
	"testing"

	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/stretchr/testify/assert"
)

func TestOpenClose(t *testing.T) {
	os.Remove("kiki.sqlite3.db")
	db, err := Open()
	assert.Nil(t, err)
	err = db.Close()
	assert.Nil(t, err)
}

func TestKeyStore(t *testing.T) {
	os.Remove("kiki.sqlite3.db")
	type A struct {
		B int
		C string
	}
	a := A{
		B: 3,
		C: "hi",
	}
	db, err := Open()
	assert.Nil(t, err)
	defer db.Close()
	err = db.Set("Astuff", "a", a)
	assert.Nil(t, err)
	var a2 A
	err = db.Get("Astuff", "a", &a2)
	assert.Nil(t, err)
	assert.Equal(t, a, a2)
}
func TestAddGetLetter(t *testing.T) {
	os.Remove("kiki.sqlite3.db")

	l := letter.Letter{
		Purpose: "share-text",
		Content: "hello, world",
	}
	sender := keypair.New()
	region := keypair.New()
	e, err := l.Seal(sender, region)
	assert.Nil(t, err)

	db, err := Open()
	assert.Nil(t, err)
	defer db.Close()
	err = db.AddEnvelope(e)
	assert.Nil(t, err)
	err = db.AddEnvelope(e)
	assert.Nil(t, err)

	e2, err := db.GetEnvelopeFromID(e.ID)
	assert.Nil(t, err)
	assert.Equal(t, e.ID, e2.ID)
	assert.Equal(t, e.Letter.Content, e2.Letter.Content)
	assert.Equal(t, e.SealedRecipients, e2.SealedRecipients)
}
