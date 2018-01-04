package letter

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/purpose"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	zack := keypair.New()
	bob := keypair.New()
	jane := keypair.New()
	donald := keypair.New()
	regionKey := keypair.New()

	l := Letter{
		To:      []string{bob.Public, jane.Public},
		Purpose: purpose.ShareText,
		Content: "hello, bob and jane",
	}

	e, err := l.Seal(zack, regionKey)
	assert.Nil(t, err)
	eBytes, err := json.Marshal(e)
	assert.Nil(t, err)

	ioutil.WriteFile("sealed.json", eBytes, 0644)

	// test validation against the region key
	err = e.Validate(regionKey)
	assert.Nil(t, err)
	err = e.Validate(donald)
	assert.NotNil(t, err)

	// test unsealing against sender
	ue, err := e.Unseal([]keypair.KeyPair{zack}, regionKey)
	assert.Nil(t, err)
	eBytes, err = json.Marshal(ue)
	assert.Nil(t, err)
	ioutil.WriteFile("unsealed.json", eBytes, 0644)

	// test against recipients
	_, err = e.Unseal([]keypair.KeyPair{bob}, regionKey)
	assert.Nil(t, err)
	_, err = e.Unseal([]keypair.KeyPair{jane}, regionKey)
	assert.Nil(t, err)

	// test against non-recipients
	_, err = e.Unseal([]keypair.KeyPair{donald}, regionKey)
	assert.NotNil(t, err)
	_, err = e.Unseal([]keypair.KeyPair{regionKey}, regionKey)
	assert.NotNil(t, err)

}
