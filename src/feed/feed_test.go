package feed

import (
	"testing"

	// "github.com/schollz/kiki/src/logging"

	"github.com/stretchr/testify/assert"
)

var f *Feed

// BenchmarkGetUser-4         	     100	  17057282 ns/op
// BenchmarkShowFeed-4        	      20	  91682205 ns/op
// BenchmarkGetBasicPosts-4   	    2000	    666736 ns/op

func init() {
	var err error
	f, err = New("testdb", "GoAabW4QeCcyeeDWZxu9wFaPAoWhbrwvrFM83JToWk33", "6ptaZoSaepphHTqQyCBRBBRF3WyKGoahXUUTVTL5BAQ3", false)
	if err != nil {
		panic(err)
	}
}

func TestGetUser(t *testing.T) {
	u := f.GetUser()
	assert.Equal(t, "8cjeQPadXXCTGe9WbqER44CqduSHpqepX4tgAoEEFH4w", u.PublicKey)
}

func BenchmarkGetUser(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f.GetUser()
	}
}

func BenchmarkShowFeed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f.ShowFeed(ShowFeedParameters{})
	}
}

func BenchmarkGetBasicPosts(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f.db.GetBasicPosts()
	}
}
