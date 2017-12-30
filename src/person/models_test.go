package person

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	p, err := New()
	assert.Nil(t, err)
	bP, err := json.Marshal(p)
	assert.Nil(t, err)
	fmt.Println(string(bP))

	var p2 *Person
	err = json.Unmarshal(bP, &p2)
	assert.Nil(t, err)
	assert.Equal(t, p, p2)
}
