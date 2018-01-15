package feed

import (
	"testing"

	"github.com/schollz/kiki/src/logging"
	"github.com/stretchr/testify/assert"
)

func TestMakePost(t *testing.T) {
	logging.Debug(false)
	f, err := Open(".")
	f.Debug(false)
	assert.Nil(t, err)
	u := f.GetUser()
	assert.Equal(t, "5z_8ZHf6cnZnortmafG0gsSX0Dl5jaOdCHUNoQiI5h8=", u.PublicKey)
}

func BenchmarkGetUser(b *testing.B) {
	f, err := Open(".")
	if err != nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		f.GetUser()
	}
}
