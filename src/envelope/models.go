package envelope

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/logging"

	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/person"
	"github.com/schollz/kiki/src/symmetric"
)

// Envelope is the sealed letter to be transfered among carriers
type Envelope struct {
	// Sealed envelope information

	// ID is the hash of the Marshaled Letter + the Public key of Sender
	ID string `json:"id",storm:"id"`
	// Timestamp is the time at which the envelope was created
	Timestamp time.Time `json:"timestamp",storm:"index"`
	// Sender is public key of the sender
	Sender *person.Person `json:"sender", storm:"index"`
	// Signature is the public key of the sender encrypted by
	// the Sender private key, against the public Region key
	// to authenticate sender. I.e., Sender == Decrypt(Signature) must be true.
	// A valid Signature is also used to prevent the Envelope from being deleted.
	// I.e., if a Region key is not able to decrypt it, then it is meant for another Region
	// and would be deleted.
	Signature string `json:"signature"`
	// Recipients is list of encypted passphrase (used to encrypt the Content)
	// encrypted against each of the public keys of the recipients.
	Recipients []string `json:"recipients"`
	// SealedLetter contains the encryoted and compressed letter,
	// encoded as base64 string
	SealedLetter string `json:"sealed_letter,omitempty"`

	// Unsealed envelope information
	// When the letter is opened, this variables will be filled. When the
	// envelope is transfered these variables should be cleared

	// Letter is the unsealed letter. Once a Envelope is "unsealed", then this
	// variable is set and the SealedLetter is set to "" (deleted). This will
	// then be saved in a bucket for unsealed letters. When the letter remains
	// sealed then this Letter is set to nil.
	Letter *letter.Letter `json:"letter,omitempty"`
	// Opened is a variable set to true if the Letter is opened, to make
	// it easier to index the opened/unopened letters in the database.false
	Opened bool `json:"opened"`
	// DeterminedRecipients are the list of recipients, decoded after unsealing a letter
	DeterminedRecipients []string `json:"determined_recipients,omitempty"`
}

// Seal will remove any identifying or confidential information
func (e *Envelope) Seal() {
	e.Letter = new(letter.Letter)
	e.Opened = false
	e.DeterminedRecipients = []string{}
}

// New creates an envelope and seals it for the specified recipients
func New(l *letter.Letter, sender *person.Person, recipients []*person.Person, regionkey *person.Person) (e *Envelope, err error) {
	logging.Log.Info("creating letter")
	e = new(Envelope)

	// Create ID (hash of any public key + hash of any content)
	h := sha256.New()
	h.Write([]byte(sender.Public()))
	h.Write([]byte(l.Base64Image))
	h.Write([]byte(l.Text))
	h.Write([]byte(l.AssignmentValue))
	h.Write([]byte(l.Replaces))
	h.Write([]byte(l.ReplyTo))
	e.ID = fmt.Sprintf("%x", h.Sum(nil))
	e.Timestamp = time.Now()
	e.Sender = sender.PublicKey()

	// Generate a passphrase to encrypt the letter
	contentBytes, err := json.Marshal(l)
	if err != nil {
		return
	}
	encryptedLetter, secretKey, err := symmetric.CompressAndEncryptWithRandomSecret(contentBytes)
	if err != nil {
		return
	}
	e.SealedLetter = base64.URLEncoding.EncodeToString(encryptedLetter)

	// the sender should always be open their own letter, so they should
	// always be a recipient
	recipients = append(recipients, sender)
	e.Recipients = make([]string, len(recipients))

	// For each recipient, generate a key-encrypted passphrase
	for i, recipient := range recipients {
		encryptedSecret, err2 := sender.Keys.Encrypt(secretKey[:], recipient.Keys)
		if err2 != nil {
			err = err2
			return
		}
		e.Recipients[i] = base64.URLEncoding.EncodeToString(encryptedSecret)
	}
	e.Opened = false

	// sign the letter by encrypting the public key against the region key
	signatureEncrypted, err := sender.Keys.Encrypt([]byte(sender.Public()), regionkey.Keys)
	if err != nil {
		return
	}
	e.Signature = base64.URLEncoding.EncodeToString(signatureEncrypted)

	return
}

// Unseal will determine the content of the letter using the identities provided
func (e *Envelope) Unseal(keysToTry []*person.Person, regionKey *person.Person) (err error) {
	err = e.unseal(keysToTry, regionKey)
	if err != nil {
		e.Seal()
	}
	return
}

func (e *Envelope) unseal(keysToTry []*person.Person, regionKey *person.Person) (err error) {
	// First validate the letter
	err = e.Validate(regionKey)
	if err != nil {
		return
	}

	var secretPassphrase [32]byte
	foundPassphrase := false
	for _, keyToTry := range keysToTry {
		for _, recipient := range e.Recipients {
			var err2 error
			encryptedSecret, err2 := base64.URLEncoding.DecodeString(recipient)
			if err2 != nil {
				err = errors.Wrap(err2, "recipients are corrupted")
				return
			}
			decryptedSecretPassphrase, err := keyToTry.Keys.Decrypt(encryptedSecret, e.Sender.Keys)
			if err == nil {
				foundPassphrase = true
				// add the known recipient to the list
				e.DeterminedRecipients = append(e.DeterminedRecipients, keyToTry.Public())
				copy(secretPassphrase[:], decryptedSecretPassphrase[:32])
			}
		}
	}
	if !foundPassphrase {
		err = errors.New("not a recipient")
		return
	}

	encryptedContent, err := base64.URLEncoding.DecodeString(e.SealedLetter)
	if err != nil {
		err = errors.Wrap(err, "content is corrupted")
		return
	}
	decrypted, err := symmetric.DecryptAndDecompress(encryptedContent, secretPassphrase)
	if err != nil {
		err = errors.Wrap(err, "content is not decryptable, corrupted?")
		return
	}
	err = json.Unmarshal(decrypted, &e.Letter)
	if err != nil {
		err = errors.Wrap(err, "problem with letter unmarshaling")
		return
	}

	e.Opened = true

	return
}

func (e *Envelope) Validate(regionKey *person.Person) (err error) {
	encryptedPublicKey, err := base64.URLEncoding.DecodeString(e.Signature)
	if err != nil {
		return
	}
	decryptedPublicKey, err := regionKey.Keys.Decrypt(encryptedPublicKey, e.Sender.Keys)
	if err != nil {
		return
	}
	if string(decryptedPublicKey) != e.Sender.Public() {
		return errors.New("validation failed, not equal")
	}
	return
}
