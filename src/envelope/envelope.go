package envelope

import (
	"encoding/base64"
	"time"

	"github.com/schollz/projectx/keypair"
	"github.com/schollz/projectx/symmetric"
)

type Envelope struct {
	Sender     string    // public key of the sender
	Recipients []string  // secret passphrase encrypted by each recipient public key
	Encrypted  string    // encrypted compressed Letter
	Timestamp  time.Time // time
	ID         string    // hash of Timestamp and Payload
}

func New(msg []byte, sender keypair.KeyPair, recipients []keyPair.KeyPair) (env *Envelope, err error) {
	env = new(Envelope)
	env.Timestamp = time.Now()
	env.ID = "1"
	encryptedBytes, secret, err := symmetric.EncryptWithRandomSecret(msg)
	if err != nil {
		return
	}
	env.Encrypted = base64.URLEncoding.EncodeToString(encryptedBytes)
	for _, recipient := range recipients {
		sender.Encrypt(secret[:], recipient)
	}
	return
}
