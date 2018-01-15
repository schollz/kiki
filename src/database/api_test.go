package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkGetPosts(b *testing.B) {
	api := Setup(".", "kiki.db")
	for i := 0; i < b.N; i++ {
		api.GetBasicPosts()
	}
}
func BenchmarkGetPosts2(b *testing.B) {
	api := Setup(".", "kiki.db")
	for i := 0; i < b.N; i++ {
		api.GetBasicPosts2()
	}
}

func TestGetPosts(t *testing.T) {
	api := Setup(".", "kiki.db")
	e, err := api.GetBasicPosts()
	assert.Nil(t, err)
	e2, err := api.GetBasicPosts2()
	assert.Nil(t, err)
	fmt.Println(len(e), len(e2))
	assert.Equal(t, len(e2), len(e))
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
