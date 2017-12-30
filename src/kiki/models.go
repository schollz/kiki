package kiki

import (
	"time"

	"github.com/schollz/kiki/src/keypair"
)

// Person is just a set of keys
type Person struct {
	keys keypair.KeyPair
}

type Envelope struct {
	Sender        string    // public key of the sender
	Recipients    []string  // secret passphrase encrypted by each recipient public key
	SealedContent string    // encrypted compressed Letter
	Timestamp     time.Time // time of entry
	ID            string    // hash of SealedContent
}

type Letter struct {
	LatestID string   // hash of sender + data
	ID       string   // original ID, different than LatestID if overwriting
	Channels []string // channels for showing the post
	ReplyTo  string   // hash that Letter is response to
	Content  LetterContent
}

type LetterContent struct {
	Data   string // base64 encoded bytes of data
	Action string // action verb
	Kind   string // kind of action
}
