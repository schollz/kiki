package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	logger.SetLevel("error")
}
func BenchmarkOpening(b *testing.B) {
	for i := 0; i < b.N; i++ {
		db, err := open("kiki.db")
		if err != nil {
			panic(err)
		}
		db.Close()
	}
}

func BenchmarkGetPosts(b *testing.B) {
	api := Setup(".", "kiki.db")
	for i := 0; i < b.N; i++ {
		api.GetBasicPosts()
	}
}
func BenchmarkGetIDs(b *testing.B) {
	api := Setup(".", "kiki.db")
	for i := 0; i < b.N; i++ {
		api.GetIDs()
	}
}

func TestGetVersions(t *testing.T) {
	api := Setup(".", "kiki.db")
	s, err := api.GetAllVersions("alskdjflkasjdf")
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(s))
	ids, err := api.GetIDs()
	assert.Nil(t, err)
	assert.True(t, len(ids) > 0)
	for id := range ids {
		s, err = api.GetAllVersions(id)
		assert.Nil(t, err)
		assert.True(t, len(s) > 0)
		break
	}
}

func TestGettingPosts(t *testing.T) {
	api := Setup(".", "kiki.db")
	e, err := api.GetBasicPosts()
	assert.Nil(t, err)
	assert.True(t, len(e) > 0)
}

func TestOpenClose(t *testing.T) {
	os.Remove("kiki.sqlite3.db")
	db, err := open("kiki.sqlite3.db")
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
	db, err := open("kiki.sqlite3.db")
	assert.Nil(t, err)
	defer db.Close()
	err = db.Set("Astuff", "a", a)
	assert.Nil(t, err)
	var a2 A
	err = db.Get("Astuff", "a", &a2)
	assert.Nil(t, err)
	assert.Equal(t, a, a2)
}
