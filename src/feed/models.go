package feed

import (
	"github.com/schollz/kiki/src/envelope"
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
//    data="new name"
//
// Send friend key:
// 		kind="send-friend-key"
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
type Message struct {
	Data        string   `json:"data"`
	Kind        string   `json:"kind" binding:"required"`
	ForPublic   bool     `json:"public"`     // use public key
	ForFriends  bool     `json:"friends"`    // use friends key
	ForSpecific []string `json:"recipients"` // list of specific people to send to
	ReplyTo     string   `json:"reply_to"`
}

// Post will generate a new letter with a message, seal it, and add it to the database.
func (p Message) Post() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "message.Post",
	})

	// make letter
	l, err := letter.New(p.Kind, p.Data, personalKey.Public())
	if err != nil {
		return
	}

	// if post, do some special functions
	if p.Kind == "post" {
		// TODO: Capture images from post
		// update l.Content.Data
		// TODO: Capture channels from post
		// update l.Channels
		// Check if its a reply to
		if p.ReplyTo != "" {
			l.ReplyTo = p.ReplyTo
		}
	} else if p.Kind == "send-friend-key" {
		// TODO: Put all friends keys into l.Content.Data
	} else if p.Kind == "assign-name" {
		// TODO: Strip HTML
	} else if p.Kind == "like" {
		// likes are public
		p.ForPublic = true
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

	// seal envelope
	logger.Debug("sealing envelope")
	e, err := envelope.New(l, personalKey, recipients)
	if err != nil {
		return
	}

	// add envelope to database
	logger.Debug("putting in carrier")
	err = db.AddEnvelope(e)
	return
}
