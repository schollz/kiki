package database

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

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

// GetLatestEnvelopeFromID returns a single envelope from its ID, trying to find the latest version of it
func (api DatabaseAPI) GetLatestEnvelopeFromID(id string) (e letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	es, err := db.getAllVersions(id)
	if err != nil {
		return
	}
	db.Close()
	return api.GetEnvelopeFromID(es[0])
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

// GetReplies returns all envelopes that are replies to a specific envelope
func (api DatabaseAPI) GetReplies(id string) (e []letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	// purpose should be to share text
	// can be empty
	// should not be replaced
	// should be a reply
	// ordered by time ascending
	ids, err := db.getAllVersions(id)
	if err != nil {
		return
	}
	envelopes, err := db.getAllFromPreparedQuery(fmt.Sprintf("SELECT * FROM letters WHERE letter_purpose = 'share-text' AND letter_replyto IN ('%s') ORDER BY time", strings.Join(ids, "','")))
	if err != nil {
		return
	}
	e = make([]letter.Envelope, len(envelopes))
	i := 0
	for _, envelope := range envelopes {
		yes, _ := db.isReplaced(envelope.ID)
		if yes {
			continue
		}
		e[i] = envelope
		i++
	}
	e = e[:i]
	return
}

func (api DatabaseAPI) GetBasicPosts() (e []letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	// purpose should be to share text
	// should not be empty
	// should not be replaced
	// should not be a reply
	return db.getAllFromPreparedQuery("SELECT * FROM letters WHERE letter_purpose = 'share-text' AND letter_content != '' AND id NOT IN (SELECT letter_replaces FROM letters WHERE letter_replaces != '') AND letter_replyto == '' ORDER BY time DESC")
}

func (self DatabaseAPI) GetBasicPosts2() (e []letter.Envelope, err error) {
	var envelopes []letter.Envelope

	db, err := open(self.FileName)
	if nil != err {
		return envelopes, err
	}
	defer db.Close()

	// query := `
	// 	SELECT
	// 	    json_array(
	// 	        json_object(
	// 	            'id', id,
	// 	            'time', time,
	// 	            'sender', sender,
	// 	            'signature', signature,
	// 	            'sealed_recipients', sealed_recipients,
	// 	            'sealed_letter', sealed_letter,
	// 	            'opened', opened,
	// 	            'letter_purpose', letter_purpose,
	// 	            'letter_to', letter_to,
	// 	            'letter_content', letter_content,
	// 	            'letter_replaces', letter_replaces,
	// 	            'letter_replyto', letter_replyto,
	// 	            'comments', json_array()
	// 	        )
	// 	    )
	// 	FROM letters
	// 	WHERE
	// 	        letter_purpose = 'share-text'
	// 	    AND
	// 	        letter_content != ''
	// 	    AND
	// 	        id NOT IN (
	// 	            SELECT letter_replaces FROM letters WHERE letter_replaces != ''
	// 	        )
	// 	    AND letter_replyto == ''
	// 	ORDER BY time DESC;
	// `

	// json1 needs to be loaded...
	query := `
		SELECT
		    '['||
		        '{'||
		            '"id": "' ||  id ||'",'||
		            '"time":"' ||  time ||'",'||
		            '"sender_raw": "' ||  sender ||'",'||
		            '"signature":"' ||  signature ||'",'||
		            '"sealed_recipients":' ||  sealed_recipients ||','||
		            '"sealed_letter":"' ||  sealed_letter ||'",'||
		            '"opened":' ||
						CASE opened
							WHEN 0 then 'false'
							ELSE 'true'
						END
					||','||
		            '"letter_purpose":"' ||  letter_purpose ||'",'||
		            '"letter_to": ' ||  letter_to ||','||
		            '"letter_content": "' ||  replace(letter_content, '"',  '''') ||'",'||
		            '"letter_replaces": "' ||  letter_replaces ||'",'||
		            '"letter_replyto": "' ||  letter_replyto ||'",'||
		            '"comments":[]'
		        ||'}'
		    ||']'
		FROM letters
		WHERE
		        letter_purpose = 'share-text'
		    AND
		        letter_content != ''
		    AND
		        id NOT IN (
		            SELECT letter_replaces FROM letters WHERE letter_replaces != ''
		        )
		    AND letter_replyto == ''
		ORDER BY time DESC;
`

	// prepare statement
	stmt, err := db.db.Prepare(query)
	if nil != err {
		return envelopes, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if nil != err {
		return envelopes, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)

		text = strings.Replace(text, "\n", "", -1)
		// err = rows.Scan(&envelopes)

		if nil != err {
			return envelopes, err
		}

		// fmt.Println(text)

		err = json.Unmarshal([]byte(text), &envelopes)
		if nil != err {
			return envelopes, err
		}
	}

	return envelopes, nil
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
	yes, err = db.isReplaced(id)
	if err != nil {
		log.Error(err)
	}
	return
}

// DiskSpaceForUser returns the bytes used by a user for recipients + sealed_content
func (api DatabaseAPI) DiskSpaceForUser(user string) (diskSpace int64, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.diskSpaceForUser(user)
}

// ListUsers returns the bytes used by a user for recipients + sealed_content
func (api DatabaseAPI) ListUsers() (users []string, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.listUsers()
}

// GetAllVersions returns the bytes used by a user for recipients + sealed_content
func (api DatabaseAPI) GetAllVersions(id string) (ids []string, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.getAllVersions(id)
}

// NumberOfLikes returns the number of likes for a post
func (api DatabaseAPI) NumberOfLikes(postID string) (likes int64) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	likes, err = db.numLikesPerPost(postID)
	if err != nil {
		log.Warn(err)
	}
	return
}

// Friends will return the followers, following and friends for a given user
func (api DatabaseAPI) Friends(publicKey string) (followers, following, friends []string) {
	followers = []string{}
	following = []string{}
	friends = []string{}
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	followers, err = db.getFollowers(publicKey)
	if err != nil {
		log.Warn(err)
	}
	following, err = db.getFollowing(publicKey)
	if err != nil {
		log.Warn(err)
	}
	followingMap := make(map[string]struct{})
	for _, f := range following {
		followingMap[f] = struct{}{}
	}
	followerMap := make(map[string]struct{})
	for _, f := range followers {
		followerMap[f] = struct{}{}
	}

	friends = make([]string, len(following)+len(followers))
	i := 0
	for _, follower := range followers {
		if _, ok := followingMap[follower]; ok {
			friends[i] = follower
			i++
		}
	}
	friends = friends[:i]

	i = 0
	for _, follower := range followers {
		if follower == "" {
			continue
		}
		if _, ok := followingMap[follower]; !ok {
			followers[i] = follower
			i++
		}
	}
	followers = followers[:i]

	i = 0
	for _, followin := range following {
		if followin == "" {
			continue
		}
		if _, ok := followerMap[followin]; !ok {
			following[i] = followin
			i++
		}
	}
	following = following[:i]
	return
}

// GetLatestKeyForFriends will return the latest key for encrypting messages to friends
func (api DatabaseAPI) GetLatestKeyForFriends(publicKey string) (key keypair.KeyPair, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.getKeyForFriends(publicKey)
}

// DeleteUsersOldestPost will delete the users oldest post
func (api DatabaseAPI) DeleteUsersOldestPost(publicKey string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.deleteUsersOldestPost(publicKey)
}
