package multibox

import (
	crypto_rand "crypto/rand"
	"testing"

	"golang.org/x/crypto/nacl/box"

	"github.com/stretchr/testify/assert"
)

func TestKeyPairs(t *testing.T) {
	senderPublicKey, senderPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}

	recipientPublicKey, recipientPrivateKey, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}
	enc, err := encryptWithKeyPair([]byte(`hello world`), senderPrivateKey, recipientPublicKey)
	assert.Nil(t, err)
	dec, err := decryptWithKeyPair(enc, senderPublicKey, recipientPrivateKey)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", string(dec))
}

func Test1(t *testing.T) {
	enc, key, err := encryptSymmetricWithRandomSecret([]byte(`hello world`))
	assert.Nil(t, err)
	dec, err := decryptSymmetric(enc, key)
	assert.Nil(t, err)
	assert.Equal(t, "hello world", string(dec))
}
