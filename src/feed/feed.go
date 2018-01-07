package feed

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/purpose"

	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/logging"
)

var (
	// public variables
	DataFolder   = "."
	DatabaseName = "kiki.db"
	IdentityFile = ""
	SettingsFile = ""
	RegionKey    keypair.KeyPair
	Port         = "8003"

	// private variables
	settings    Settings
	personalKey keypair.KeyPair
	db          *database.Database
	log         = logging.Log
)

// Setup initializes the kiki instance
func Setup() (err error) {
	// define region key
	RegionKey, err = keypair.FromPair("rbcDfDMIe8qXq4QPtIUtuEylDvlGynx56QgeHUZUZBk=",
		"GQf6ZbBbnVGhiHZ_IqRv0AlfqQh1iofmSyFOcp1ti8Q=") // define region key
	if err != nil {
		return
	}

	// Setup database
	log.Debug("setting up database")
	database.Setup(path.Join(DataFolder, DatabaseName))
	err = database.Set("AssignedNames", RegionKey.Public, "Public")
	if err != nil {
		return
	}

	// Define personalKey for this instance
	log.Debug("setting up personalKey")
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
	log.Debug("setting up settings")
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

	return
}

// ProcessLetter will determine where to put the letter
func ProcessLetter(l letter.Letter) (err error) {
	if !purpose.Valid(l.Purpose) {
		err = errors.New("invalid purpose")
		return
	}

	// rewrite the letter.To array so that it contains
	// public keys that are valid
	newTo := []string{}
	for _, to := range l.To {
		switch to {
		case "public":
			newTo = append(newTo, RegionKey.Public)
		case "friends":
			friendsKeyPairs, err2 := database.GetKeysFromSender(personalKey.Public)
			if err2 != nil {
				return err2
			}
			for _, friendsKeyPair := range friendsKeyPairs {
				newTo = append(newTo, friendsKeyPair.Public)
			}
		default:
			_, err2 := keypair.FromPublic(to)
			if err2 != nil {
				log.Infof("Not a valid public key: '%s'", to)
			} else {
				newTo = append(newTo, to)
			}
		}
	}
	l.To = newTo

	// seal the letter
	e, err := l.Seal(personalKey, RegionKey)
	if err != nil {
		return
	}

	err = database.AddEnvelope(e)
	if err != nil {
		return
	}

	err = UnsealLetters()
	return
}

// UnsealLetters will go through unopened envelopes and open them
// and then add them to the database.
func UnsealLetters() (err error) {
	envelopes, err := database.GetAllEnvelopes(false)
	if err != nil {
		return err
	}
	keysToTry := []keypair.KeyPair{personalKey, RegionKey}
	for _, envelope := range envelopes {
		ue, err := envelope.Unseal(keysToTry, RegionKey)
		if err != nil {
			log.Debug(err)
			continue
		}
		log.Debug(ue.Letter)
		err = database.AddEnvelope(ue)
		if err != nil {
			return err
		}
	}
	return
}

// NewPerson will generate a new person, and a friends key.
// It will automatically post the new friends key to your feed.
func NewPerson() (p keypair.KeyPair, err error) {
	// generate a new person
	p = keypair.New()
	if err != nil {
		return
	}

	// generate a key for friends
	myfriends := keypair.New()
	if err != nil {
		return
	}
	myfriendsByte, err := json.Marshal(myfriends)

	// share the friends key with yourself
	l := letter.Letter{
		Purpose: purpose.ShareKey,
		Content: string(myfriendsByte),
	}
	e, err := l.Seal(p, RegionKey)
	if err != nil {
		return
	}

	// post the envelope
	err = database.AddEnvelope(e)
	return
}

// // UpdateFriendsKeys will prepend the Friends key determine from envelopes, if
// // is not already added.
// func UpdateFriendsKeys(e *envelope.Envelope) (err error) {
// 	logger := logging.Log.WithFields(logrus.Fields{
// 		"func": "UpdateFriendsKeys",
// 	})

// 	var newKey *person.Person
// 	err = json.Unmarshal([]byte(e.Letter.Text), &newKey)

// 	keyBucket := "keysFromFriends"
// 	if e.Sender.Public() == personalKey.Public() {
// 		// key was sent from someone other than you
// 		keyBucket = "keysForFriends"
// 	}

// 	var friendKeys []*person.Person
// 	err = db.Get("keystore", keyBucket, friendKeys)
// 	if err != nil {
// 		friendKeys = []*person.Person{}
// 	}
// 	for _, key := range friendKeys {
// 		if key == newKey {
// 			return nil
// 		}
// 	}
// 	friendKeys = append([]*person.Person{newKey}, friendKeys...)
// 	err = db.Set("keystore", keyBucket, friendKeys)
// 	log.Debugf("new %s: '%s' sent %v", keyBucket, newKey.Public(), utils.TimeAgo(e.Timestamp))
// 	return
// }

// // UpdateNames will prepend the Friends key determine from envelopes, if
// // is not already added.
// func UpdateNames(e *envelope.Envelope) (err error) {
// 	logger := logging.Log.WithFields(logrus.Fields{
// 		"func": "UpdateNames",
// 	})

// 	err = db.Set("AssignedNames", e.Sender.Public(), e.Letter.Text)
// 	if err != nil {
// 		return
// 	}
// 	log.Debugf("public name '%s' -> '%s' (%s)", e.Sender.Public()[:8], e.Letter.Text, utils.TimeAgo(e.Timestamp))
// 	return
// }
