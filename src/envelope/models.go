package envelope

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/keypair"

	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/person"
	"github.com/schollz/kiki/src/symmetric"
)

// Envelope is the sealed letter to be transfered among carriers
type Envelope struct {
	Sender     *keypair.KeyPair `json:"sender"`     // public key of the sender
	Recipients []string         `json:"recipients"` // secret passphrase to open SealedContent,
	// encrypted by each recipient public key
	SealedContent string    `json:"sealed_content"` // encrypted compressed Letter
	Timestamp     time.Time `json:"timestamp"`      // time of entry
	ID            string    `json:"id"`             // hash of SealedContent
	content       *letter.Letter
}

// New creates an envelope and seals it for the specified recipients
func New(l *letter.Letter, sender *person.Person, recipients []*person.Person) (e *Envelope, err error) {
	e = new(Envelope)
	e.ID = l.ID
	e.Timestamp = time.Now()
	e.Sender = sender.Keys.PublicKey()

	contentBytes, err := json.Marshal(l)
	if err != nil {
		return
	}
	encryptedLetter, secretKey, err := symmetric.CompressAndEncryptWithRandomSecret(contentBytes)
	if err != nil {
		return
	}
	e.SealedContent = base64.StdEncoding.EncodeToString(encryptedLetter)

	recipients = append(recipients, sender) // the sender should always be open their own letter
	e.Recipients = make([]string, len(recipients))
	for i, recipient := range recipients {
		encryptedSecret, err2 := sender.Keys.Encrypt(secretKey[:], recipient.Keys)
		if err2 != nil {
			err = err2
			return
		}
		e.Recipients[i] = base64.StdEncoding.EncodeToString(encryptedSecret)
	}

	return
}

func (e *Envelope) Unseal(opener *person.Person) (err error) {
	var secretPassphrase [32]byte
	foundPassphrase := false
	for _, recipient := range e.Recipients {
		var err2 error
		encryptedSecret, err2 := base64.StdEncoding.DecodeString(recipient)
		if err2 != nil {
			err = errors.Wrap(err2, "recipients are corrupted")
			return
		}
		decryptedSecretPassphrase, err := opener.Keys.Decrypt(encryptedSecret, e.Sender)
		if err == nil {
			foundPassphrase = true
			copy(secretPassphrase[:], decryptedSecretPassphrase[:32])
			break
		}
	}
	if !foundPassphrase {
		err = errors.New("not a recipient")
		return
	}

	encryptedContent, err := base64.StdEncoding.DecodeString(e.SealedContent)
	if err != nil {
		err = errors.Wrap(err, "content is corrupted")
		return
	}
	decrypted, err := symmetric.DecryptAndDecompress(encryptedContent, secretPassphrase)
	if err != nil {
		err = errors.Wrap(err, "content is not decryptable, corrupted?")
		return
	}
	err = json.Unmarshal(decrypted, &e.content)
	if err != nil {
		err = errors.Wrap(err, "problem with letter unmarshaling")
		return
	}
	return
}
