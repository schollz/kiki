package feed

import (
	"errors"

	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/person"
	"github.com/sirupsen/logrus"
)

type Settings struct {
	StoragePerPublicPerson int64 // maximum size in bytes to store of public messages. Once exceeded, old messages are purged
	FriendsOfFriends       bool  // whether you want to share your friends friend keys with new friends, effectively making a new friend friends with all your friends. This also means that when you make a new friend, that friends key is emitted to all your current friends. (default: true)
	ShowPublicPhotos       bool  // if true, automatically show the display public photos (default: false)
}

// GenerateSettings create new instance of Something
func GenerateSettings() Settings {
	return Settings{
		StoragePerPublicPerson: 5000000, // 5 MB
		FriendsOfFriends:       true,
		ShowPublicPhotos:       true,
	}
}

// Message
// ## TYPES OF MESSAGES ##
//
// Post new message:
//		data="HTML of post"
// 		kind="post"
// 		public=true/false (optional, default:false)
// 		friends=true/false (optional, default:false)
// 		recipients=["which people"] (optional, default:nil)
//
// Post new photo:
//		data="base64 of photo data"
// 		kind=".jpg"/".png"
// 		public=true/false (optional, default:false)
// 		friends=true/false (optional, default:false)
// 		recipients=["which people"] (optional, default:nil)
//
// Assign name:
// 		kind="assign-name"
//    data="plaintext of name"
//
// Assign public profile:
// 		kind="assign-profile"
//    data="HTML of profile"
//
// Assign public profile image:
//		kind="assign-image"
//		data="base64 of photo data"
//
// Send friend key:
// 		kind="give-key"cd
//    recipients=["friend1 public key","friend2 public key"]
//
// "like" something:
//		kind="like"
//    data="post ID"
//
// Follow someone:
//		kind="follow"
//		recipients=["user ID"]
//
// Ghost someone:
//		kind="ghost"
//		data=["user ID"]
//
type Message struct {
	Data        string   `json:"data"`
	Kind        string   `json:"kind" binding:"required"`
	ForPublic   bool     `json:"public"`     // use public key
	ForFriends  bool     `json:"friends"`    // use friends key
	ForSpecific []string `json:"recipients"` // list of specific people to send to
	ReplyTo     string   `json:"reply_to"`
}

type LetterPost struct {
	Letter     *letter.Letter
	Recipients []string
	ForFriends bool
	ForPublic  bool
}

func (p LetterPost) Post() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "message.Post",
	})

	logger.Infof("posting %v", p)

	// make letter
	l := new(letter.Letter)

	// if post, do some special functions
	switch p.Letter.Kind {
	case "post-text":
		// TODO: Capture images from post
		// update l.Content.Data
		// post the images as new messages
		// TODO: Capture channels from post
		// update l.Channels
		// Check if its a reply to
	case "give-key":
		// TODO: Put all friends keys into l.Content.Data
	case "assign-name":
		// TODO: Strip HTML
		// assigned names are public
		p.ForPublic = true
	case "assign-profile":
		// TODO: Strip images
		// profiles are public
		p.ForPublic = true
	case "assign-image":
		// profile images are public
		p.ForPublic = true
	case ".jpg":
		// do nothing
	case ".png":
		// do nothing
	case "like":
		// likes are public
		p.ForPublic = true
	case "follow":
		// follows are public?
		p.ForPublic = true
	case "ghost":
		// blocks are private
		p.ForPublic = false
		p.ForFriends = false
		// TODO: issue new friends key
		// emit new friends key to all remaining friends
		// (TODO: in feed, when encountering a ghost in Unsealed Envelopes, remove the public key specified in the data)
	default:
		return errors.New("message kind not supported: " + p.Kind)
	}

	if p.ReplyTo != "" {
		l.RepliesTo(p.ReplyTo)
	}

	// determine recipients
	// _Note:_ the current sender is automatically added when sealing the envelope.
	recipients := []*person.Person{}
	if p.ForPublic {
		//  Add public key
		recipients = append(recipients, RegionKey)
	}
	if p.ForFriends {
		// TODO: Add friends key
	}
	for _, pubString := range p.ForSpecific {
		otherRecipient, err := person.FromPublicKey(pubString)
		if err != nil {
			logging.Log.Infof("not a valid public key: '%s'", pubString)
			continue
		}
		recipients = append(recipients, otherRecipient)
	}

	// TODO: seal envelope

	// TODO: add envelope to database
}

// Post will generate a new letter with a message, seal it, and add it to the database.
func (p Message) Post() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "message.Post",
	})

	logger.Infof("posting %v", p)

	// make letter
	l := new(letter.Letter)

	// if post, do some special functions
	switch p.Kind {
	case "post-text":
		// TODO: Capture images from post
		// update l.Content.Data
		// post the images as new messages
		// TODO: Capture channels from post
		// update l.Channels
		// Check if its a reply to
	case "give-key":
		// TODO: Put all friends keys into l.Content.Data
	case "assign-name":
		// TODO: Strip HTML
		// assigned names are public
		p.ForPublic = true
	case "assign-profile":
		// TODO: Strip images
		// profiles are public
		p.ForPublic = true
	case "assign-image":
		// profile images are public
		p.ForPublic = true
	case ".jpg":
		// do nothing
	case ".png":
		// do nothing
	case "like":
		// likes are public
		p.ForPublic = true
	case "follow":
		// follows are public?
		p.ForPublic = true
	case "ghost":
		// blocks are private
		p.ForPublic = false
		p.ForFriends = false
		// TODO: issue new friends key
		// emit new friends key to all remaining friends
		// (TODO: in feed, when encountering a ghost in Unsealed Envelopes, remove the public key specified in the data)
	default:
		return errors.New("message kind not supported: " + p.Kind)
	}

	if p.ReplyTo != "" {
		l.RepliesTo(p.ReplyTo)
	}

	// determine recipients
	// _Note:_ the current sender is automatically added when sealing the envelope.
	recipients := []*person.Person{}
	if p.ForPublic {
		//  Add public key
		recipients = append(recipients, RegionKey)
	}
	if p.ForFriends {
		// TODO: Add friends key
	}
	for _, pubString := range p.ForSpecific {
		otherRecipient, err := person.FromPublicKey(pubString)
		if err != nil {
			logging.Log.Infof("not a valid public key: '%s'", pubString)
			continue
		}
		recipients = append(recipients, otherRecipient)
	}

	// TODO: seal envelope

	// TODO: add envelope to database
	return
}
