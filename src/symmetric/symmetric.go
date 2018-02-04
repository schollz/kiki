package symmetric

import (
	crypto_rand "crypto/rand"
	"errors"
	"io"

	"github.com/schollz/kiki/src/utils"
	"golang.org/x/crypto/nacl/secretbox"
)

func CompressAndEncryptWithRandomSecret(msg []byte) (encrypted []byte, secretKey [32]byte, err error) {
	return EncryptWithRandomSecret(utils.CompressByte(msg))
}

func EncryptWithRandomSecret(msg []byte) (encrypted []byte, secretKey [32]byte, err error) {
	if _, err = io.ReadFull(crypto_rand.Reader, secretKey[:]); err != nil {
		return
	}

	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err = io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return
	}

	// This encrypts msg and appends the result to the nonce.
	encrypted = secretbox.Seal(nonce[:], msg, &nonce, &secretKey)
	return
}

func Decrypt(encrypted []byte, secretKey [32]byte) (decrypted []byte, err error) {
	// When you decrypt, you must use the same nonce and key you used to
	// encrypt the message. One way to achieve this is to store the nonce
	// alongside the encrypted message. Above, we stored the nonce in the first
	// 24 bytes of the encrypted text.
	var decryptNonce [24]byte
	copy(decryptNonce[:], encrypted[:24])
	decrypted, ok := secretbox.Open(nil, encrypted[24:], &decryptNonce, &secretKey)
	if !ok {
		err = errors.New("decryption failed")
	}
	return
}

func DecryptAndDecompress(encrypted []byte, secretKey [32]byte) (decrypted []byte, err error) {
	decrypted, err = Decrypt(encrypted, secretKey)
	decrypted = utils.DecompressByte(decrypted)
	return
}
