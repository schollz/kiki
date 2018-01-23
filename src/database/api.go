package database

import (
	// "encoding/json"
	"fmt"
	"path"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/purpose"
)

// Publicly acessible database routines
type DatabaseAPI struct {
	FileName string
}

func Setup(locationToDatabase string, databaseName ...string) (api DatabaseAPI) {
	name := "kiki.db"
	if len(databaseName) > 0 {
		name = databaseName[0]
	}
	api = DatabaseAPI{
		FileName: path.Join(locationToDatabase, name),
	}
	fmt.Printf("setup database at '%s'\n", api.FileName)
	return
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

// AddTags will add the tags to the database
func (api DatabaseAPI) AddTags(idToTags map[string][]string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	for id := range idToTags {
		for _, tag := range idToTags[id] {
			err = db.AddTag(tag, id)
			if err != nil {
				return
			}
		}
	}
	return
}

// GetEnvelopesFromTag
func (api DatabaseAPI) GetEnvelopesFromTag(tag string) (es []letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	// ids, err := db.GetIDsFromTag(tag)
	// if err != nil {
	// 	return
	// }

	// es, err = db.getAllFromPreparedQuery(fmt.Sprintf("SELECT * FROM (SELECT * FROM letters WHERE opened ==1 AND letter_purpose = 'share-text' AND letter_content != '' AND id IN ('%s') AND letter_replyto == '' ORDER BY TIME) GROUP BY letter_firstid ORDER BY time DESC;", strings.Join(ids, "','")))
	es, err = db.getAllFromPreparedQuery(fmt.Sprintf(`
		SELECT * FROM
			(
				SELECT
					*
				FROM
					letters
				INNER JOIN tags
		            ON tags.tag ='%s'
		            AND tags.e_id = letters.id
				WHERE
					opened ==1
				AND
					letter_purpose = 'share-text'
				AND
					letter_content != ''
				AND
					letter_replyto == ''
				ORDER BY time
			) GROUP BY letter_firstid ORDER BY time DESC;
		`, tag))
	return
}

func (api DatabaseAPI) AddEnvelope(e letter.Envelope) (err error) {
	_, err = api.GetEnvelopeFromID(e.ID)
	if err == nil {
		return errors.New("envelope already exists")
	}
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.addEnvelope(e)
}

func (api DatabaseAPI) UpdateEnvelope(e letter.Envelope) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.addEnvelope(e)
}

// GetEnvelopeFromID returns a single envelope from its ID and returns an error if it does not exist.
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
		if len(es) > 0 {
			e = es[0]
		} else {
			err = errors.New("envelope does not exist")
		}
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
		db.Close()
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
	envelopes, err := db.getAllFromPreparedQuery(fmt.Sprintf("SELECT * FROM (SELECT * FROM letters WHERE opened == 1 AND letter_purpose = '%s' AND letter_replyto == ? ORDER BY time) GROUP BY letter_firstid ORDER BY time", purpose.ShareText), id)
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
	// should not be replaced (GROUP BY letter_firstid)
	// should not be a reply
	es, err := db.getAllFromPreparedQuery("SELECT * FROM (SELECT * FROM letters WHERE opened ==1 AND letter_purpose = 'share-text' AND letter_replyto == '' ORDER BY time) GROUP BY letter_firstid ORDER BY time DESC")
	if err != nil {
		return
	}
	i := 0
	e = make([]letter.Envelope, len(es))
	for _, es0 := range es {
		if es0.Letter.Content == "" {
			continue
		}
		e[i] = es0
		i++
	}
	e = e[:i]
	return
}

func (api DatabaseAPI) GetBasicPostsForUser(publickey string) (e []letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	// purpose should be to share text
	// should not be empty
	// should not be replaced
	// should not be a reply
	es, err := db.getAllFromPreparedQuery("SELECT * FROM (SELECT * FROM letters WHERE opened ==1 AND letter_purpose = 'share-text' AND sender == ? AND letter_replyto == '' ORDER BY time) GROUP BY letter_firstid ORDER BY time DESC;", publickey)
	if err != nil {
		return
	}
	i := 0
	e = make([]letter.Envelope, len(es))
	for _, es0 := range es {
		if es0.Letter.Content == "" {
			continue
		}
		e[i] = es0
		i++
	}
	e = e[:i]
	return
}

// GetBasicPostLatest returns the latest post for a person
func (api DatabaseAPI) GetBasicPostLatest(publickey string) (e letter.Envelope, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	// purpose should be to share text
	// should not be empty
	// should not be replaced
	// should not be a reply
	es, err := db.getAllFromPreparedQuery("SELECT * FROM (SELECT * FROM letters WHERE opened == 1 AND letter_purpose = 'share-text' AND sender == ? ORDER BY time DESC) GROUP BY letter_firstid ORDER BY time LIMIT 1;", publickey)
	if err != nil {
		return
	}
	if len(es) == 0 {
		err = errors.New("no posts")
	} else {
		e = es[0]
	}
	e = es[0]
	return
}

func (self DatabaseAPI) jsonFormatting(payload string) string {
	payload = strings.Replace(payload, "\n", "", -1)
	payload = strings.Replace(payload, "\"null\"", "null", -1)
	payload = strings.Replace(payload, ",]", "]", -1)
	return payload
}

func (self DatabaseAPI) postJsonSql() string {
	return `
		'{'||
			'"id": "' ||  letter_firstid ||'",'||
			'"timestamp": ' || strftime('%s',time) ||','||
			'"recipients": ' ||  letter_to ||','||
			'"owner_id": "' ||  sender ||'",'||
			'"owner_name": "' || IFNULL((SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == 'action-assign/name' AND sender == ltr.sender ORDER BY time DESC LIMIT 1), 'null') ||'",'||
			'"content": "' ||  replace(letter_content, '"',  '''') ||'",'||
			'"reply_to": "' ||  letter_replyto ||'",'||
			'"purpose":"' ||  letter_purpose ||'",'||
			'"likes": '|| (SELECT COUNT(*) FROM letters WHERE opened == 1 AND letter_purpose == 'action-like' AND letter_content=ltr.letter_firstid) ||','||
			'"num_comments": '|| ( SELECT count(*) FROM letters WHERE opened == 1 AND letter_purpose = 'share-text' AND letter_replyto = ltr.letter_firstid )
		||'}'
	`
}

// json1 needs to be loaded...
func (self DatabaseAPI) GetPostsForApi() ([]ApiBasicPost, error) {
	var posts []ApiBasicPost

	logger.Log.Info(self.FileName)

	db, err := open(self.FileName)
	if nil != err {
		return posts, err
	}
	defer db.Close()

	query := `
		SELECT
	        ` + self.postJsonSql() + `
		FROM letters AS ltr
		WHERE
				opened == 1
			AND
		        letter_purpose = 'share-text'
		    AND
				letter_replyto == ''
		GROUP BY letter_firstid
		ORDER BY time DESC;
`

	// prepare statement
	stmt, err := db.db.Prepare(query)
	if nil != err {
		return posts, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if nil != err {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)

		if nil != err {
			return posts, err
		}

		text = self.jsonFormatting(text)

		var post ApiBasicPost
		if err = post.Unmarshal(text); nil != err {
			return posts, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (self DatabaseAPI) GetPostCommentsForApi(post_id string) ([]ApiBasicPost, error) {
	var posts []ApiBasicPost

	db, err := open(self.FileName)
	if nil != err {
		return posts, err
	}
	defer db.Close()

	query := `
		SELECT
			` + self.postJsonSql() + `
		FROM letters AS ltr
		WHERE
				opened == 1
			AND
		        letter_purpose = 'share-text'
		    AND letter_replyto == ?
		GROUP BY letter_firstid
		ORDER BY time DESC;
`

	// prepare statement
	stmt, err := db.db.Prepare(query)
	if nil != err {
		return posts, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(post_id)
	if nil != err {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)
		if nil != err {
			return posts, err
		}

		text = self.jsonFormatting(text)

		var post ApiBasicPost
		if err = post.Unmarshal(text); nil != err {
			return posts, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (self DatabaseAPI) GetPostVersionsForApi(post_id string) ([]ApiBasicPost, error) {
	var posts []ApiBasicPost

	db, err := open(self.FileName)
	if nil != err {
		return posts, err
	}
	defer db.Close()

	query := `
		SELECT
			` + self.postJsonSql() + `
		FROM letters AS ltr
		WHERE
				opened == 1
			AND
		        letter_purpose = 'share-text'
		    AND letter_firstid == ?
		ORDER BY time DESC;
`

	// prepare statement
	stmt, err := db.db.Prepare(query)
	if nil != err {
		return posts, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(post_id)
	if nil != err {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)
		if nil != err {
			return posts, err
		}

		text = self.jsonFormatting(text)

		var post ApiBasicPost
		if err = post.Unmarshal(text); nil != err {
			return posts, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (self DatabaseAPI) GetPostForApi(post_id string) ([]ApiBasicPost, error) {
	var posts []ApiBasicPost

	db, err := open(self.FileName)
	if nil != err {
		return posts, err
	}
	defer db.Close()

	query := `
		SELECT
			` + self.postJsonSql() + `
		FROM letters AS ltr
		WHERE
				opened == 1
			AND
				letter_purpose = 'share-text'
			AND
				letter_firstid = ?
		ORDER BY time DESC LIMIT 1;
`

	// prepare statement
	stmt, err := db.db.Prepare(query)
	if nil != err {
		return posts, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(post_id)
	if nil != err {
		return posts, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)
		if nil != err {
			return posts, err
		}

		text = self.jsonFormatting(text)

		var post ApiBasicPost
		if err = post.Unmarshal(text); nil != err {
			return posts, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func (self DatabaseAPI) GetUserForApi(user_id string) (ApiUser, error) {
	var user ApiUser

	db, err := open(self.FileName)
	if nil != err {
		return user, err
	}
	defer db.Close()

	query := `
		SELECT
	        '{'||
	            '"public_key": "' ||  ? ||'",'||
				'"name": "' || IFNULL((SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == 'action-assign/name' AND sender == ? ORDER BY time DESC LIMIT 1), 'null') ||'",'||
				'"profile": "' || IFNULL((SELECT replace(letter_content, '"',  '''') FROM letters WHERE opened == 1 AND letter_purpose == 'action-assign/profile' AND sender == ? ORDER BY time DESC LIMIT 1), 'null') ||'",'||
				'"image": "' || IFNULL((SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == 'action-assign/image' AND sender == ? ORDER BY time DESC LIMIT 1), 'null') ||'",'||
				'"followers": [' || (
					SELECT IFNULL(GROUP_CONCAT(ids), '') FROM (
						SELECT IFNULL('"'||sender||'"', '') AS ids FROM letters WHERE letter_purpose = 'action-follow' AND letter_content = ?
					)
				) ||'],'||
				'"following": [' || (
					SELECT IFNULL(GROUP_CONCAT(ids), '') FROM (
						SELECT IFNULL('"'||letter_content||'"', '') AS ids FROM letters WHERE letter_purpose = 'action-follow' AND sender = ?
					)
				) ||'],'||
				'"blocked": [' || (
					SELECT IFNULL(GROUP_CONCAT(ids), '') FROM (
						SELECT IFNULL('"'||letter_content||'"', '') AS ids FROM letters WHERE letter_purpose = 'action-block' AND sender = ?
					)
				) ||'],'||
				'"friends": ['|| (
		            SELECT IFNULL(GROUP_CONCAT(ids), '') FROM (
		                SELECT '"'||sender||'"' AS ids FROM letters WHERE letter_purpose = 'action-follow' AND letter_content = ?
		                INTERSECT
		                SELECT '"'||letter_content||'"' AS ids FROM letters WHERE letter_purpose = 'action-follow' AND sender = ?
		            )
		        ) ||']'
			||'}';
`

	// prepare statement
	stmt, err := db.db.Prepare(query)
	if nil != err {
		return user, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(user_id, user_id, user_id, user_id, user_id, user_id, user_id, user_id, user_id)
	if nil != err {
		return user, err
	}
	defer rows.Close()

	for rows.Next() {
		var text string
		err = rows.Scan(&text)
		if nil != err {
			return user, err
		}

		text = self.jsonFormatting(text)
		if err = user.Unmarshal(text); nil != err {
			return user, err
		}
	}

	return user, err
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

// RemoveLettersForUser will delete the envelopes for a specific user
func (api DatabaseAPI) RemoveLettersForUser(user string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	err = db.deleteLettersFromSender(user)
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

// ListBlockedUsers returns the bytes used by a user for recipients + sealed_content
func (api DatabaseAPI) ListBlockedUsers(publickey string) (users []string, err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.listBlockedUsers(publickey)
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

// DeleteUsersEdits will delete the users edits made to posts
func (api DatabaseAPI) DeleteUsersEdits(publicKey string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.deleteUsersEdits(publicKey)
}

// DeleteOldActions will delete all the old actions of a user
func (api DatabaseAPI) DeleteOldActions(publicKey string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	err = db.deleteUsersOldActions(publicKey, purpose.ActionName)
	if err != nil {
		return
	}
	err = db.deleteUsersOldActions(publicKey, purpose.ActionProfile)
	if err != nil {
		return
	}
	err = db.deleteUsersOldActions(publicKey, purpose.ActionImage)
	if err != nil {
		return
	}
	return
}

// DeleteProfiles will delete everything for all users that have submitted an action-erase
func (api DatabaseAPI) DeleteProfiles() (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.deleteUsers()
}

// DeleteUser will delete everything for all users that have submitted an action-erase
func (api DatabaseAPI) DeleteUser(publicKey string) (err error) {
	db, err := open(api.FileName)
	if err != nil {
		return
	}
	defer db.Close()
	return db.deleteUser(publicKey)
}
