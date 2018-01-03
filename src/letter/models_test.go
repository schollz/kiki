package letter

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	l := NewText("this is some text!")

	lB, err := json.Marshal(l)
	assert.Nil(t, err)

	fmt.Println(string(lB))
}
