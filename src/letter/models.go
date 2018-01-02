package letter

import (
	"crypto/sha256"
	"fmt"
)

// Letter specifies the content being transfered to the self, or other users.
// The Letter can contain "posts", images or text. These are content generated
// a user. They are specified by the Kind "post-image" or "post-text".
// A Letter is also used to specify public data about an Assignment. These are
// things assigned to users (follows, profile names, profile images), or posts
// (channels, likes). An Assignment is public because it should be used to
// quantify the reputation of the object it is assignet.
type Letter struct {
	// Kind specifies the kind of letter. Currently the kinds are:
	// "assign-X" - used to assign public data for reputation purposes (likes, follows, channel subscriptions, settting profile images and text and names)
	// "post-X" - used to post either text or images
	Kind     string `json:"kind"` // kind of letter
	Replaces string `json:"replaces"`

	// "post-text"
	// Channels is a list of the channels to put the letter in
	Channels []string `json:"channels,omitempty"`
	// ReplyTo is the ID of the post being responded to
	ReplyTo string `json:"reply_to,omitempty"`
	// HTML is processed by stripping images and re-posting them as their own
	HTML string `json:"html,omitempty"`
	// Plaintext is processed to strip tags
	Plaintext string `json:"plaintext,omitempty"`

	// for "post-image"
	// Extension for an image is either "jpg" or "png" ("gif" not supported)
	Extension string `json:"extension",omitempty`
	// Base64Image is a base64 encoded data of image
	Base64Image string `json:"base64_image",omitempty"`

	// for "assign-X"
	// AssignmentValue is the value going to be inserted for the assignment.
	// Assignment works by modifying a bucket in the keystore. The bucket is
	// the assignment type, "X" (could be "name", "profile", etc.). The key in
	// in the bucket is the public key of the sender (determined from envelope).
	// The value inserted is the AssignmentValue, determine for each assignment as
	// name: Plaintext containing name
	// profile: HTML containing profile
	// profileimage: ID of the image
	// like: ID of post
	// channel: Name of channel
	// follow: ID of person to follow
	AssignmentValue string `json:"assignment_value",omitempty"`
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
