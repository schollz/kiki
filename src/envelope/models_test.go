package envelope

import (
	"encoding/json"
	"fmt"
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

	u, err := e.Unseal([]*person.Person{zack})
	assert.Nil(t, err)
	assert.Equal(t, u.Letter.Content.Data, "hello, bob and jane")
	fmt.Println(u)

	u, err = e.Unseal([]*person.Person{bob})
	assert.Nil(t, err)
	u, err = e.Unseal([]*person.Person{jane})
	assert.Nil(t, err)
	u, err = e.Unseal([]*person.Person{donald})
	assert.NotNil(t, err)

	bE, _ := json.Marshal(e)
	ioutil.WriteFile("e.json", bE, 0644)

	myPeople := []*person.Person{donald, bob, jane, zack}
	u, err = e.Unseal(myPeople)
	for _, p := range myPeople {
		fmt.Println(p.Keys.Public)
	}
	for _, p := range u.Recipients {
		fmt.Println(p.Public)
	}
}
