package database

import (
	"path"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
)

// Publicly acessible database routines
type DatabaseAPI struct {
	FileName string
}

func Setup(locationToDatabase string) (api DatabaseAPI) {
	return DatabaseAPI{
		FileName: path.Join(locationToDatabase, "kiki.sqlite3.db"),
	}
}

func (api DatabaseAPI) Set(bucket, key string, value interface{}) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.Set(bucket, key, value)
}

func (api DatabaseAPI) Get(bucket, key string, value interface{}) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.Get(bucket, key, value)
}

func (api DatabaseAPI) AddEnvelope(e letter.Envelope) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.addEnvelope(e)
}

// GetEnvelopeFromID returns a single envelope from its ID
func (api DatabaseAPI) GetEnvelopeFromID(id string) (e letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	var es []letter.Envelope
	es, err = db.getAllFromPreparedQuery("SELECT * FROM letters WHERE id = ?", id)
	if err != nil {
		err = errors.Wrap(err, "GetEnvelopeFromID("+id+")")
	} else {
		e = es[0]
	}
	return
}

// GetAllEnvelopes returns all envelopes determined by whether they are opened
func (api DatabaseAPI) GetAllEnvelopes(opened ...bool) (e []letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	if len(opened) > 0 {
		if opened[0] {
			return db.getAllFromQuery("SELECT * FROM letters WHERE opened == 1 ORDER BY time DESC")
		} else {
			return db.getAllFromQuery("SELECT * FROM letters WHERE opened == 0 ORDER BY time DESC")
		}
	} else {
		return db.getAllFromQuery("SELECT * FROM letters ORDER BY time DESC")
	}
}

// GetKeys will return all the keys
func (api DatabaseAPI) GetKeys() (s []keypair.KeyPair, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.getKeys()
}

// GetKeysFromSender will return all the keys from a certain sender
func (api DatabaseAPI) GetKeysFromSender(sender string) (s []keypair.KeyPair, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.getKeys(sender)
}

// GetName will return the assigned name for the public key of a sender
func (api DatabaseAPI) GetName(publicKey string) (name string) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	name, err = db.getName(publicKey)
	if err != nil {
		log.Warn(err)
	}
	return
}

// GetProfile will return the assigned profile for the public key of a sender
func (api DatabaseAPI) GetProfile(publicKey string) (name string) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	name, err = db.getProfile(publicKey)
	if err != nil {
		log.Warn(err)
	}
	return
}

// GetProfile will return the assigned profile for the public key of a sender
func (api DatabaseAPI) GetProfileImage(publicKey string) (imageID string) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	imageID, err = db.getProfileImage(publicKey)
	if err != nil {
		log.Warn(err)
	}
	return
}

// GetUser returns information for a user
func (api DatabaseAPI) GetUser(publicKey string) (name, profile, image string) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	name, _ = db.getName(publicKey)
	profile, _ = db.getProfile(publicKey)
	image, _ = db.getProfileImage(publicKey)
	return
}

// GetFriendsName will search friend's keys and determine the name of the friends key, e.g. Zack's Friends (where Zack is assigned name of public key)
func (api DatabaseAPI) GetFriendsName(publicKey string) (name string) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.getFriendsName(publicKey)
}

// RemoveLetters will delete the letter containing that ID
func (api DatabaseAPI) RemoveLetters(ids []string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	for _, id := range ids {
		err2 := db.deleteLetterFromID(id)
		if err2 != nil {
			log.Warn(err2)
		}
	}
	return
}

// GetIDs will delete the letter containing that ID
func (api DatabaseAPI) GetIDs() (ids map[string]struct{}, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	s, err := db.getIDs()
	if err != nil {
		return
	}
	ids = make(map[string]struct{})
	for _, id := range s {
		ids[id] = struct{}{}
	}
	return
}

// IsReplaced returns boolean of whether post with ID has been replaced
func (api DatabaseAPI) IsReplaced(id string) (yes bool) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.isReplaced(id)
}
