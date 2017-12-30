package letter

import (
	"crypto/sha256"
	"fmt"
)

// Letter contains meta data describing the content
type Letter struct {
	LatestID string        `json:"latest_id"` // hash of sender + un-encrypted data
	ID       string        `json:"id"`        // original ID, different than LatestID if overwriting
	Channels []string      `json:"channels"`  // channels for showing the post
	ReplyTo  string        `json:"reply_to"`  // hash that Letter is response to
	Content  LetterContent `json:"content"`
}

// LetterContent is the actual content of the letter
type LetterContent struct {
	Kind string `json:"kind"` // kind of letter content
	Data string `json:"data"` // base64 encoded bytes of data
}

func New(kind, data, publicKey string) (l *Letter, err error) {
	l = new(Letter)
	l.Content = LetterContent{
		Kind: kind,
		Data: data,
	}
	h := sha256.New()
	h.Write([]byte(publicKey))
	h.Write([]byte(data))
	l.ID = fmt.Sprintf("%x", h.Sum(nil))
	l.LatestID = l.ID
	l.Channels = []string{}
	l.ReplyTo = ""
	return
}
