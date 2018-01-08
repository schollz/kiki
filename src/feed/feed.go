package feed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/purpose"

	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/utils"
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
		// generate a new personal key
		personalKey = keypair.New()
		pBytes, err2 := json.Marshal(personalKey)
		if err2 != nil {
			return err2
		}
		err2 = ioutil.WriteFile(IdentityFile, pBytes, 0644)
		if err2 != nil {
			return err2
		}

		// add the friends key
		err2 = AddFriendsKey()
		if err2 != nil {
			return err2
		}

		// block the region public key from being used as a sender, ever
		err2 = ProcessLetter(letter.Letter{
			To:      []string{"public"},
			Purpose: purpose.AssignBlock,
			Content: RegionKey.Public,
		})
		if err2 != nil {
			err2 = errors.Wrap(err2, "setup")
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
		case "self":
			// automatically done when adding any letter
			// this just put here for pedantic reasons
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
	if personalKey == RegionKey {
		err = errors.New("cannot post with region key")
		return
	}
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

// UnsealLetters will go through unopened envelopes and open them and then add them to the database. Also go through and purge bad letters (invalidated letters)
func UnsealLetters() (err error) {
	lettersToPurge := []string{}
	envelopes, err := database.GetAllEnvelopes(false)
	if err != nil {
		return err
	}
	keysToTry := []keypair.KeyPair{personalKey, RegionKey}
	for _, envelope := range envelopes {
		if err := envelope.Validate(RegionKey); err != nil {
			// add to purge
			lettersToPurge = append(lettersToPurge, envelope.ID)
		}
		ue, err := envelope.Unseal(keysToTry, RegionKey)
		if err != nil {
			log.Debug(err)
			continue
		}
		log.Debug(ue.Letter)
		err = database.AddEnvelope(ue)
		if err != nil {
			log.Debug(err)
			continue
		}
	}

	// purge invalid letters
	if len(lettersToPurge) > 0 {
		err = database.RemoveLetters(lettersToPurge)
	}
	return
}

func ShowFeed() (err error) {
	envelopes, err := database.GetAllEnvelopes(true)
	if err != nil {
		return
	}
	log.Debugf("Found %d envelopes", len(envelopes))
	for _, e := range envelopes {
		if e.Letter.Purpose != purpose.ShareText {
			continue
		}
		senderName, err2 := database.GetName(e.Sender.Public)
		if err2 != nil {
			log.Warn(err2)
			senderName = e.Sender.Public
		}
		fmt.Printf("%s (%s) [%s]:\n%s\n\n", senderName, e.Sender.Public, utils.TimeAgo(e.Timestamp), e.Letter.Content)

	}
	return
}

// AddFriendsKey will generate a new friends key and post it to the feed
func AddFriendsKey() (err error) {
	// generate a key for friends
	myfriends := keypair.New()
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey")
		return
	}
	myfriendsByte, err := json.Marshal(myfriends)

	// share the friends key with yourself
	err = ProcessLetter(letter.Letter{
		To:      []string{"self"},
		Purpose: purpose.ShareKey,
		Content: string(myfriendsByte),
	})
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey")
		return
	}

	// block the friends public key from being used as a sender, ever
	err = ProcessLetter(letter.Letter{
		To:      []string{"public"},
		Purpose: purpose.AssignBlock,
		Content: myfriends.Public,
	})
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey")
		return
	}

	return
}
