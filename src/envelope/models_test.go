package envelope

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/person"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	zack, _ := person.New()
	bob, _ := person.New()
	jane, _ := person.New()
	donald, _ := person.New()

	l, _ := letter.New("post", "hello, bob and jane", zack.Keys.Public)
	e, err := New(l, zack, []*person.Person{bob, jane})
	assert.Nil(t, err)

	_, err = json.Marshal(e)
	assert.Nil(t, err)

	err = e.Unseal(zack)
	assert.Nil(t, err)
	assert.Equal(t, e.content.Content.Data, "hello, bob and jane")

	err = e.Unseal(bob)
	assert.Nil(t, err)
	err = e.Unseal(jane)
	assert.Nil(t, err)
	err = e.Unseal(donald)
	assert.NotNil(t, err)

	bE, _ := json.Marshal(e)
	ioutil.WriteFile("e.json", bE, 0644)
}
