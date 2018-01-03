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
	regionKey, _ := person.New()

	l := letter.NewText("hello, bob and jane")
	e, err := New(l, zack, []*person.Person{bob, jane}, regionKey)
	assert.Nil(t, err)

	_, err = json.Marshal(e)
	assert.Nil(t, err)

	err = e.Unseal([]*person.Person{zack}, regionKey)
	assert.Nil(t, err)
	assert.Equal(t, e.Letter.Text, "hello, bob and jane")

	err = e.Unseal([]*person.Person{bob}, regionKey)
	assert.Nil(t, err)
	err = e.Unseal([]*person.Person{jane}, regionKey)
	assert.Nil(t, err)
	err = e.Unseal([]*person.Person{donald}, regionKey)
	assert.NotNil(t, err)

	bE, _ := json.Marshal(e)
	ioutil.WriteFile("e.json", bE, 0644)

	myPeople := []*person.Person{donald, bob, jane, zack}
	err = e.Unseal(myPeople, regionKey)
	for _, p := range myPeople {
		fmt.Println(p.Keys.Public)
	}
	for _, p := range e.DeterminedRecipients {
		fmt.Println(p)
	}

	eBytes, _ := json.Marshal(e)
	fmt.Println(string(eBytes))

	err = e.Validate(regionKey)
	assert.Nil(t, err)
	err = e.Validate(donald)
	assert.NotNil(t, err)
}
