package kiki

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/envelope"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/person"
	"github.com/schollz/kiki/src/utils"
	"github.com/sirupsen/logrus"
)

var (
	DataFolder   = "."
	DatabaseName = "kiki.db"
	IdentityFile = path.Join(DataFolder, "identity.json")
	RegionKey    *person.Person
	Identity     *person.Person
	Port         = "8003"
	db           *database.Database
)

// Setup initializes the kiki instance
func Setup() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "kiki-Setup()",
	})

	// define region key
	RegionKey, err = person.FromPublicPrivateKeys("rbcDfDMIe8qXq4QPtIUtuEylDvlGynx56QgeHUZUZBk=",
		"GQf6ZbBbnVGhiHZ_IqRv0AlfqQh1iofmSyFOcp1ti8Q=") // define region key
	if err != nil {
		return
	}

	// Setup database
	logger.Debug("setting up database")
	db = database.Setup(path.Join(DataFolder, DatabaseName))
	db.Set("AssignedNames", RegionKey.Public(), "Public")

	// Setup identity for this instance
	logger.Debug("setting up identity")
	if _, err := os.Stat(IdentityFile); os.IsNotExist(err) {
		var err2 error
		p, err2 := NewPerson()
		if err2 != nil {
			return err2
		}
		pBytes, err2 := json.Marshal(p)
		if err2 != nil {
			return err2
		}
		err2 = ioutil.WriteFile(IdentityFile, pBytes, 0644)
		if err2 != nil {
			return err2
		}
	}
	pBytes, err := ioutil.ReadFile(IdentityFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(pBytes, &Identity)
	if err != nil {
		return
	}

	err = RegenerateFeed()
	return
}

// PostMessage will generate a new letter with a message, seal it, and add it to the database.
func PostMessage(kind, message string, isPublic bool, recipients ...*person.Person) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "PostMessage",
	})

	// check if recipients is available
	if len(recipients) == 0 {
		recipients = []*person.Person{}
	}

	// make letter
	if len(message) < 10 {
		logger.Debugf("new %s: '%s'", kind, message)
	} else {
		logger.Debugf("new %s: '%s'", kind, message[:10])
	}
	l, err := letter.New(kind, message, Identity.Public())
	if err != nil {
		return
	}

	// add Region key if it is public
	if isPublic {
		recipients = append(recipients, RegionKey)
	}

	// seal envelope
	logger.Debug("sealing envelope")
	e, err := envelope.New(l, Identity, recipients) // the current sender is automatically added
	if err != nil {
		return
	}

	// add envelope to database
	logger.Debug("putting in carrier")
	err = db.AddEnvelope(e)
	return
}

// OpenEnvelopes will process a JSON marshaled byte of a person
// to open any of the sealed envelopes that have not been opened.
func OpenEnvelopes() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "OpenEnvelopes",
	})

	// get all the envelopes
	envelopes, err := db.GetEnvelopes()
	if err != nil {
		return
	}
	logging.Log.Debugf("found %d envelopes", len(envelopes))
	for _, e := range envelopes {
		// see if its already been done
		_, errGet := db.GetUnsealedEnvelope(e.ID)
		if errGet == nil {
			logger.Debugf("skipping '%s..', already opened", e.ID[:6])
			continue
		}

		// unseal
		ue, err := e.Unseal([]*person.Person{Identity, RegionKey})
		if err != nil {
			continue // this letter is not for this person
		}

		// add unsealed letter to database
		errAdd := db.AddUnsealedEnvelope(ue)
		if errAdd != nil {
			return errors.Wrap(errAdd, "problem opening")
		}
	}
	return
}

// NewPerson will generate a new person, and a friends key.
// It will automatically post the new friends key to your feed.
func NewPerson() (p *person.Person, err error) {
	// generate a new person
	p, err = person.New()
	if err != nil {
		return
	}

	// generate a key for friends
	myfriends, err := person.New()
	if err != nil {
		return
	}
	myfriendsByte, err := json.Marshal(myfriends)

	// post the key to yourself
	e, err := envelope.SelfAddress(p, "friends-key", string(myfriendsByte))
	if err != nil {
		return
	}

	// post the envelope
	err = db.AddEnvelope(e)
	return
}

func ShowMessages() (err error) {
	envelopes, err := db.GetUnsealedEnvelopes()
	if err != nil {
		return
	}
	for _, e := range envelopes {
		if e.Letter.Content.Kind == "post" {
			var userName string
			db.Get("AssignedNames", e.Sender.Public, &userName)
			recipientNames := make([]string, len(e.Recipients))
			for i, recipient := range e.Recipients {
				var name string
				db.Get("AssignedNames", recipient.Public, &name)
				if name == "" {
					name = "?"
				}
				recipientNames[i] = name
			}
			fmt.Printf(`-----------------
%s[%s] -> %s (%s)
			
%s
`, userName, e.Sender.Public, strings.Join(recipientNames, ","), utils.TimeAgo(e.Timestamp), e.Letter.Content.Data)
		}
	}
	return
}

// RegenerateFeed will update all the parameters in the kiki instance
// by reading through the unsealed envelopes to get keys for friends,
// keys from friends, assigned names.
func RegenerateFeed() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "RegenerateFeed",
	})
	logger.Debug("starting")
	envelopes, err := db.GetUnsealedEnvelopes()
	if err != nil {
		return
	}
	for _, e := range envelopes {
		errProcess := ProcessLetter(e)
		if errProcess != nil {
			return errors.Wrap(errProcess, "problem processing "+e.ID)
		}
	}
	ShowMessages()
	return
}

// ProcessLetter will determine what to do with each letter.
func ProcessLetter(e *envelope.UnsealedEnvelope) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "ProcessLetter",
	})

	switch kind := e.Letter.Content.Kind; kind {
	case "friends-key":
		return UpdateFriendsKeys(e)
	case "assign-name":
		return UpdateNames(e)
	case "post":
		return nil
	default:
		logger.Warnf("unknown kind: '%s'", kind)
	}
	return
}

// UpdateFriendsKeys will prepend the Friends key determine from envelopes, if
// is not already added.
func UpdateFriendsKeys(e *envelope.UnsealedEnvelope) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "UpdateFriendsKeys",
	})

	var newKey *person.Person
	err = json.Unmarshal([]byte(e.Letter.Content.Data), &newKey)

	keyBucket := "keysFromFriends"
	if e.Sender.Public == Identity.Public() {
		// key was sent from someone other than you
		keyBucket = "keysForFriends"
	}

	var friendKeys []*person.Person
	err = db.Get("keystore", keyBucket, friendKeys)
	if err != nil {
		friendKeys = []*person.Person{}
	}
	for _, key := range friendKeys {
		if key == newKey {
			return nil
		}
	}
	friendKeys = append([]*person.Person{newKey}, friendKeys...)
	err = db.Set("keystore", keyBucket, friendKeys)
	logger.Debugf("new %s: '%s' sent %v", keyBucket, newKey.Public(), utils.TimeAgo(e.Timestamp))
	return
}

// UpdateNames will prepend the Friends key determine from envelopes, if
// is not already added.
func UpdateNames(e *envelope.UnsealedEnvelope) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "UpdateNames",
	})

	err = db.Set("AssignedNames", e.Sender.Public, e.Letter.Content.Data)
	if err != nil {
		return
	}
	logger.Debugf("public name '%s' -> '%s' (%s)", e.Sender.Public[:8], e.Letter.Content.Data, utils.TimeAgo(e.Timestamp))
	return
}
