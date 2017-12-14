package keypair

import (
	crypto_rand "crypto/rand"
	"testing"

	"golang.org/x/crypto/nacl/box"

	"github.com/stretchr/testify/assert"
)

func BenchmarkEncrypt(b *testing.B) {
	bob, _ := New()
	jane, _ := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bob.Encrypt([]byte(`hello, world. this, is 32 bytes!`), jane)
		if err != nil {
			panic(err)
		}
	}
}
func BenchmarkDecrypt(b *testing.B) {
	bob, _ := New()
	jane, _ := New()
	enc, err := bob.Encrypt([]byte(`hello, world. this, is 32 bytes!`), jane)
	if err != nil {
		panic(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jane.Decrypt(enc, bob)
		if err != nil {
			panic(err)
		}
	}
}
func TestKeyPairEncryption(t *testing.T) {
	bob, err := New()
	assert.Nil(t, err)
	jane, _ := New()
	enc, err := bob.Encrypt([]byte(`hello, world`), jane)
	assert.Nil(t, err)
	dec, err := jane.Decrypt(enc, bob)
	assert.Nil(t, err)
	assert.Equal(t, []byte(`hello, world`), dec)
}

func TestKeyPairs(t *testing.T) {
	sendPublicKeyString, senderPrivateKeyString := GenerateKeys()
	senderPublicKey, err := keyStringToBytes(sendPublicKeyString)
	assert.Nil(t, err)
	senderPrivateKey, _ := keyStringToBytes(senderPrivateKeyString)
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
