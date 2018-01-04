package keypair

import (
	"bytes"
	crypto_rand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/box"
)

type KeyPair struct {
	Public  string `json:"public"`
	Private string `json:"private,omitempty"`
	private *[32]byte
	public  *[32]byte
}

func New() (kp KeyPair) {
	var err error
	kp = KeyPair{}
	kp.Public, kp.Private = GenerateKeys()
	kp.public, err = keyStringToBytes(kp.Public)
	if err != nil {
		panic(err)
	}
	kp.private, err = keyStringToBytes(kp.Private)
	if err != nil {
		panic(err)
	}
	return
}

func (kp KeyPair) PublicKey() (kpPublic KeyPair) {
	kpPublic = KeyPair{}
	kpPublic.Public = kp.Public
	kpPublic.public, _ = keyStringToBytes(kpPublic.Public)
	return
}

func FromPair(public, private string) (kp KeyPair, err error) {
	kp = KeyPair{}
	kp.Public, kp.Private = public, private
	kp.public, err = keyStringToBytes(kp.Public)
	if err != nil {
		return
	}
	kp.private, err = keyStringToBytes(kp.Private)
	if err != nil {
		return
	}
	return
}

// FromPublic generates a half-key pair
func FromPublic(public string) (kp KeyPair, err error) {
	kp = KeyPair{}
	kp.Public = public
	kp.public, err = keyStringToBytes(kp.Public)
	if err != nil {
		return
	}
	return
}

func (kp KeyPair) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	if kp.Private == "" {
		buffer.WriteString(fmt.Sprintf("\"%s\":\"%s\"", "public", kp.Public))
	} else {
		buffer.WriteString(fmt.Sprintf("\"%s\":\"%s\",", "public", kp.Public))
		buffer.WriteString(fmt.Sprintf("\"%s\":\"%s\"", "private", kp.Private))
	}
	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

type KeyPairJSON struct {
	Public  string
	Private string
}

func (kp *KeyPair) UnmarshalJSON(b []byte) (err error) {
	var kpBase KeyPairJSON
	err = json.Unmarshal(b, &kpBase)
	if err != nil {
		return
	}
	kp.Public = kpBase.Public
	kp.Private = kpBase.Private
	if len(kpBase.Public) > 0 {
		kp.public, err = keyStringToBytes(kpBase.Public)
		if err != nil {
			return
		}
	}
	if len(kpBase.Private) > 0 {
		kp.private, err = keyStringToBytes(kpBase.Private)
		if err != nil {
			return
		}
	}
	return
}

func (kp KeyPair) Encrypt(msg []byte, recipient KeyPair) (encrypted []byte, err error) {
	encrypted, err = encryptWithKeyPair(msg, kp.private, recipient.public)
	return
}

func (kp KeyPair) Decrypt(encrypted []byte, sender KeyPair) (msg []byte, err error) {
	msg, err = decryptWithKeyPair(encrypted, sender.public, kp.private)
	return
}

func GenerateKeys() (publicKey, privateKey string) {
	publicKeyBytes, privateKeyBytes, err := box.GenerateKey(crypto_rand.Reader)
	if err != nil {
		panic(err)
	}

	publicKey = base64.URLEncoding.EncodeToString(publicKeyBytes[:])
	privateKey = base64.URLEncoding.EncodeToString(privateKeyBytes[:])
	return
}

func keyStringToBytes(s string) (key *[32]byte, err error) {
	keyBytes, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return
	}
	key = new([32]byte)
	copy(key[:], keyBytes[:32])
	return
}

func encryptWithKeyPair(msg []byte, senderPrivateKey, recipientPublicKey *[32]byte) (encrypted []byte, err error) {
	// You must use a different nonce for each message you encrypt with the
	// same key. Since the nonce here is 192 bits long, a random value
	// provides a sufficiently small probability of repeats.
	var nonce [24]byte
	if _, err = io.ReadFull(crypto_rand.Reader, nonce[:]); err != nil {
		return
	}
	// This encrypts msg and appends the result to the nonce.
	encrypted = box.Seal(nonce[:], msg, &nonce, recipientPublicKey, senderPrivateKey)
	return
}

func decryptWithKeyPair(enc []byte, senderPublicKey, recipientPrivateKey *[32]byte) (decrypted []byte, err error) {
	// The recipient can decrypt the message using their private key and the
	// sender's public key. When you decrypt, you must use the same nonce you
	// used to encrypt the message. One way to achieve this is to store the
	// nonce alongside the encrypted message. Above, we stored the nonce in the
	// first 24 bytes of the encrypted text.
	var decryptNonce [24]byte
	copy(decryptNonce[:], enc[:24])
	var ok bool
	decrypted, ok = box.Open(nil, enc[24:], &decryptNonce, senderPublicKey, recipientPrivateKey)
	if !ok {
		err = errors.New("keypair decryption failed")
	}
	return
}
