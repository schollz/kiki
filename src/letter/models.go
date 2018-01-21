package letter

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/mr-tron/base58/base58"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/logging"

	"github.com/schollz/kiki/src/symmetric"
)

// // TESTING
// type ApiBasicPost struct {
// 	ID          string   `json:"id"`
// 	Recipients  []string `json:"recipients"`
// 	ReplyTo     string   `json:"reply_to"`
// 	Content     string   `json:"content"`
// 	Timestamp   int64    `json:"timestamp"`
// 	OwnerId     string   `json:"owner_id"`
// 	OwnerName   string   `json:"owner_name"`
// 	Likes       int64    `json:"likes"`
// 	NumComments int64    `json:"num_comments"`
// 	Purpose     string   `json:"purpose,omitempty"`
// }
//
// type ApiUser struct {
// 	Name      string `json:"name"`
// 	PublicKey string `json:"public_key"`
// 	Profile   string `json:"profile"`
// 	Image     string `json:"image"`
// 	// Followers
// 	// Following
// 	// Friends
// }

// Letter specifies the content being transfered to the self, or other users. The Letter has a purpose - either to "share" or to do an "action". You can "share" posts  or images. You action things like follows, likes, profile names, etc.
type Letter struct {
	// Purpose specifies the purpose of letter. Currently the purposes are:
	// "action-X" - used to action public data for reputation purposes (likes, follows, channel subscriptions, settting profile images and text and names)
	// "share-X" - used to share content either "post" or "image/png"/"image/jpg"
	Purpose string `json:"purpose,omitempty"`

	// To is a list of who the letter is addressed to: "public", "friends", "self" or the public key of any person
	To []string `json:"to,omitempty"`

	// Content is is the content of the letter (base64 encoded image, text, or HTML)
	Content string `json:"content,omitempty"`

	// Replaces is the ID that this letter will replace if it is opened
	Replaces string `json:"replaces,omitempty"`

	// ReplyTo is the ID of the post being responded to
	ReplyTo string `json:"reply_to,omitempty"`
}

// Envelope is the sealed letter to be transfered among carriers
type Envelope struct {
	// Sealed envelope information

	// ID is the hash of the Marshaled Letter + the Public key of Sender
	ID string `json:"id"`
	// Timestamp is the time at which the envelope was created
	Timestamp time.Time `json:"timestamp"`
	// Sender is public key of the sender
	Sender    keypair.KeyPair `json:"sender"`
	SenderRaw string          `json:"sender_raw"`
	// Signature is the public key of the sender encrypted by
	// the Sender private key, against the public Region key
	// to authenticate sender. I.e., Sender == Decrypt(Signature) must be true.
	// A valid Signature is also used to prevent the Envelope from being deleted.
	// I.e., if a Region key is not able to decrypt it, then it is meant for another Region
	// and would be deleted.
	Signature string `json:"signature"`
	// SealedRecipients is list of encypted passphrase (used to encrypt the Content)
	// encrypted against each of the public keys of the recipients.
	SealedRecipients []string `json:"sealed_recipients"`

	// SealedLetter contains the encryoted and compressed letter, encoded as base64 string
	SealedLetter string `json:"sealed_letter,omitempty"`

	// Unsealed envelope information
	// When the letter is opened, this variables will be filled. When the
	// envelope is transfered these variables should be cleared

	// Letter is the unsealed letter. Once a Envelope is "unsealed", then this
	// variable is set and the SealedLetter is set to "" (deleted). This will
	// then be saved in a bucket for unsealed letters. When the letter remains
	// sealed then this Letter is set to nil.
	Letter Letter `json:"letter,omitempty"`

	// Opened is a variable set to true if the Letter is opened, to make
	// it easier to index the opened/unopened letters in the database.false
	Opened bool `json:"opened"`
}

func (e *Envelope) Close() {
	e.Letter = Letter{}
	e.Opened = false
}

// Seal creates an envelope and seals it for the specified recipients
func (l Letter) Seal(sender keypair.KeyPair, regionkey keypair.KeyPair) (e Envelope, err error) {
	logging.Log.Info("creating letter")

	// generate a list of keypairs for each public key of the recipients in letter.To
	newTo := make([]string, len(l.To)+1)
	recipients := make([]keypair.KeyPair, len(l.To)+1)
	recipients[0] = sender
	newTo[0] = sender.Public
	for i, publicKeyOfRecipient := range l.To {
		recipients[i+1], err = keypair.FromPublic(publicKeyOfRecipient)
		newTo[i+1] = publicKeyOfRecipient
		if err != nil {
			return
		}
	}
	l.To = newTo

	e = Envelope{}

	e.Timestamp = time.Now().UTC()
	e.Sender = sender.PublicKey()
	// Create blockchain ID (hash of any public key + hash of any content + replaces)
	h := sha256.New()
	h.Write([]byte(sender.Public))
	h.Write([]byte(l.Purpose))
	h.Write([]byte(l.Content))
	h.Write([]byte(l.Replaces))
	h.Write([]byte(l.ReplyTo))
	h.Write([]byte(strings.Join(l.To, ",")))
	e.ID = base58.FastBase58Encoding(h.Sum(nil))

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

	// For each recipient, generate a key-encrypted passphrase
	e.SealedRecipients = make([]string, len(recipients))
	for i, recipient := range recipients {
		encryptedSecret, err2 := sender.Encrypt(secretKey[:], recipient)
		if err2 != nil {
			err = err2
			return
		}
		e.SealedRecipients[i] = base64.URLEncoding.EncodeToString(encryptedSecret)
	}

	// sign the letter by encrypting the public key against the region key
	signatureEncrypted, err := sender.Encrypt([]byte(sender.Public), regionkey)
	if err != nil {
		return
	}
	e.Signature = base64.URLEncoding.EncodeToString(signatureEncrypted)

	// Remove the letter information
	e.Opened = false
	e.Letter = Letter{}

	return
}

// Unseal will determine the content of the letter using the identities provided
func (e Envelope) Unseal(keysToTry []keypair.KeyPair, regionKey keypair.KeyPair) (Envelope, error) {
	e2, err := e.unseal(keysToTry, regionKey)
	if err != nil {
		return e, err
	}
	return e2, nil
}

func (e2 Envelope) unseal(keysToTry []keypair.KeyPair, regionKey keypair.KeyPair) (e Envelope, err error) {
	e = e2
	// First validate the letter
	err = e.Validate(regionKey)
	if err != nil {
		return
	}

	var secretPassphrase [32]byte
	foundPassphrase := false
	for _, keyToTry := range keysToTry {
		for _, recipient := range e.SealedRecipients {
			var err2 error
			encryptedSecret, err2 := base64.URLEncoding.DecodeString(recipient)
			if err2 != nil {
				err = errors.Wrap(err2, "recipients are corrupted")
				return
			}
			decryptedSecretPassphrase, err := keyToTry.Decrypt(encryptedSecret, e.Sender)
			if err == nil {
				foundPassphrase = true
				copy(secretPassphrase[:], decryptedSecretPassphrase[:32])
				break
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

func (e Envelope) Validate(regionKey keypair.KeyPair) (err error) {
	if e.Sender.Public == regionKey.Public {
		return errors.New("region cannot be sender")
	}
	encryptedPublicKey, err := base64.URLEncoding.DecodeString(e.Signature)
	if err != nil {
		return
	}
	decryptedPublicKey, err := regionKey.Decrypt(encryptedPublicKey, e.Sender)
	if err != nil {
		return
	}
	if string(decryptedPublicKey) != e.Sender.Public {
		return errors.New("signature corrupted")
	}
	return
}
