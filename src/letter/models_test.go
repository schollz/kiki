package letter

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/schollz/kiki/src/person"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	p, err := person.New()
	assert.Nil(t, err)

	l, err := New("post", "this is my first **post**!", p.Keys.Public)
	assert.Nil(t, err)

	lB, err := json.Marshal(l)
	assert.Nil(t, err)

	fmt.Println(string(lB))
}
