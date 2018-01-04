package feed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/purpose"

	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/envelope"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/person"
	"github.com/schollz/kiki/src/utils"
	"github.com/sirupsen/logrus"
)

var (
	// public variables
	DataFolder   = "."
	DatabaseName = "kiki.db"
	IdentityFile = ""
	SettingsFile = ""
	RegionKey    *person.Person
	Port         = "8003"

	// private variables
	settings    Settings
	personalKey *person.Person
	db          *database.Database
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

	// Define personalKey for this instance
	logger.Debug("setting up personalKey")
	if IdentityFile == "" {
		IdentityFile = path.Join(DataFolder, "identity.json")
	}
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
	err = json.Unmarshal(pBytes, &personalKey)
	if err != nil {
		return
	}

	// Define settings for this instance
	logger.Debug("setting up settings")
	if SettingsFile == "" {
		SettingsFile = path.Join(DataFolder, "settings.json")
	}
	if _, err := os.Stat(SettingsFile); os.IsNotExist(err) {
		s := GenerateSettings()
		pBytes, err2 := json.Marshal(s)
		if err2 != nil {
			return err2
		}
		err2 = ioutil.WriteFile(SettingsFile, pBytes, 0644)
		if err2 != nil {
			return err2
		}
	}
	pBytes, err = ioutil.ReadFile(SettingsFile)
	if err != nil {
		return
	}
	err = json.Unmarshal(pBytes, &settings)
	if err != nil {
		return
	}

	err = RegenerateFeed()
	return
}

// ProcessLetter will determine where to put the letter
func (l Letter) ProcessLetter(l Letter) (err error) {
	if !purpose.Valid(l.Purpose) {
		err = errors.New("invalid purpose")
		return
	}

	// rewrite the letter.To array so that it contains
	// public keys
	newTo := []string{}
	for _, to := range l.To {
		switch to {
		case "public":
			newTo = append(newTo, RegionKey)

		}
	}
	if l.To == "public" {

	}
	return
}

// OpenEnvelopes will process a JSON marshaled byte of a person
// to open any of the sealed envelopes that have not been opened.
func OpenEnvelopes() (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "OpenEnvelopes",
	})

	logger.Info("opening envelopes")

	// get all the unopened envelopes, to be opened
	envelopes, err := db.GetEnvelopes(false)
	if err != nil {
		return
	}
	logging.Log.Debugf("found %d unopened envelopes", len(envelopes))
	for _, e := range envelopes {
		// unseal
		err := e.Unseal([]*person.Person{personalKey, RegionKey}, RegionKey)
		if err != nil {
			continue // this letter is not for this person
		}

		// add unsealed letter to database
		errAdd := db.AddEnvelope(e)
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
	l := letter.NewAssignment("assign-friend", string(myfriendsByte))
	e, err := envelope.New(l, p, []*person.Person{}, RegionKey)
	if err != nil {
		return
	}

	// post the envelope
	err = db.AddEnvelope(e)
	return
}

func ShowMessages() (err error) {
	// get the opened envelopes
	envelopes, err := db.GetEnvelopes(true)
	if err != nil {
		return
	}
	for _, e := range envelopes {
		if strings.Contains("post-", e.Letter.Kind) {
			var userName string
			db.Get("AssignedNames", e.Sender.Public, &userName)
			recipientNames := make([]string, len(e.Recipients))
			for i, recipient := range e.DeterminedRecipients {
				var name string
				db.Get("AssignedNames", recipient, &name)
				if name == "" {
					name = "?"
				}
				recipientNames[i] = name
			}
			fmt.Printf(`-----------------
%s[%s] -> %s (%s)
			
%s
`, userName, e.Sender.Public(), strings.Join(recipientNames, ","), utils.TimeAgo(e.Timestamp), e.Letter.Text)
		}
	}
	return
}

// RegenerateFeed will update all the parameters in the kiki instance
// by reading through the unsealed envelopes to get keys for friends,
// keys from friends, assigned names.
func RegenerateFeed() (err error) {
	return nil
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "RegenerateFeed",
	})
	logger.Debug("starting")
	// get all the opened envelopes
	envelopes, err := db.GetEnvelopes(true)
	if err != nil {
		return
	}
	for _, e := range envelopes {
		errProcess := processLetter(e)
		if errProcess != nil {
			return errors.Wrap(errProcess, "problem processing "+e.ID)
		}
	}
	ShowMessages()
	return
}

// processLetter will determine what to do with each letter.
func processLetter(e *envelope.Envelope) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "processLetter",
	})

	switch kind := e.Letter.Kind; kind {
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
func UpdateFriendsKeys(e *envelope.Envelope) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "UpdateFriendsKeys",
	})

	var newKey *person.Person
	err = json.Unmarshal([]byte(e.Letter.Text), &newKey)

	keyBucket := "keysFromFriends"
	if e.Sender.Public() == personalKey.Public() {
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
func UpdateNames(e *envelope.Envelope) (err error) {
	logger := logging.Log.WithFields(logrus.Fields{
		"func": "UpdateNames",
	})

	err = db.Set("AssignedNames", e.Sender.Public(), e.Letter.Text)
	if err != nil {
		return
	}
	logger.Debugf("public name '%s' -> '%s' (%s)", e.Sender.Public()[:8], e.Letter.Text, utils.TimeAgo(e.Timestamp))
	return
}
