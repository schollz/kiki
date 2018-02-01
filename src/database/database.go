package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cihub/seelog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/purpose"
	flock "github.com/theckman/go-flock"
)

var (
	logger logging.SeelogWrapper
	log    seelog.LoggerInterface
)

type database struct {
	name     string
	db       *sql.DB
	fileLock *flock.Flock
}

func init() {
	logger = logging.New()
	log = logger.Log
}

func Debug(verbose bool) {
	if verbose {
		logger.SetLevel("debug")
	} else {
		logger.SetLevel("warn")
	}
	log = logger.Log
}

// open will open the database for transactions by first aquiring a filelock.
func open(fileName string, readOnly ...bool) (d *database, err error) {
	d = new(database)
	// convert the name to base64 for file writing
	d.name = fileName

	// if read-only, make sure the database exists
	if _, err = os.Stat(d.name); err != nil && len(readOnly) > 0 && readOnly[0] {
		err = errors.New(fmt.Sprintf("database '%s' does not exist", d.name))
		return
	}

	// obtain a lock on the database
	d.fileLock = flock.NewFlock(d.name + ".lock")
	for {
		locked, err := d.fileLock.TryLock()
		if err == nil && locked {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// check if it is a new database
	newDatabase := false
	if _, err := os.Stat(d.name); os.IsNotExist(err) {
		newDatabase = true
	}

	// open sqlite3 database
	d.db, err = sql.Open("sqlite3", d.name)
	if err != nil {
		return
	}

	// create new database tables if needed
	if newDatabase {
		err = d.MakeTables()
		if err != nil {
			return
		}
	}

	return
}

// Close will close the database connection and remove the filelock.
func (d *database) Close() (err error) {
	// close filelock
	err = d.fileLock.Unlock()
	if err != nil {
		log.Error(err)
	} else {
		os.Remove(d.name + ".lock")
	}
	// close database
	err2 := d.db.Close()
	if err2 != nil {
		err = err2
		log.Error(err)
	}
	return
}

// MakeTables creates two tables, a `keystore` table:
//
// 	BUCKET_KEY (TEXT)	VALUE (TEXT)
//
// and also a `letters`:
func (d *database) MakeTables() (err error) {
	sqlStmt := `CREATE TABLE tags (tag TEXT, e_id TEXT);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}
	sqlStmt = `CREATE index tags_idx on tags(tag,e_id);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}

	sqlStmt = `create table keystore (bucket_key text not null primary key, value text);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}
	sqlStmt = `create index keystore_idx on keystore(bucket_key);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables")
		return
	}
	// The "letters" table contains all the envelopes (opened and unopened) and their respective inforamtion in the letters.
	sqlStmt = `create table letters (id text not null primary key, time TIMESTAMP, sender text, signature text, sealed_recipients text, sealed_letter text, opened integer, letter_purpose text, letter_to text, letter_content text, letter_firstid text, letter_replyto text, unique(id), UNIQUE(signature));`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables, letters")
		return
	}

	// indices
	sqlStmt = `CREATE INDEX idx_sender ON letters(opened,letter_purpose,sender,letter_content);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables, letters")
		return
	}

	// indices
	sqlStmt = `CREATE INDEX idx_content ON letters(opened,letter_purpose,letter_content,id,letter_replyto);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables, letters")
		return
	}

	// indices
	sqlStmt = `CREATE INDEX idx_replyto ON letters(opened,letter_purpose,letter_replyto);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables, letters")
		return
	}

	// indices
	sqlStmt = `CREATE INDEX idx_replaces ON letters(opened,letter_firstid);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables, letters")
		return
	}

	// indices
	sqlStmt = `CREATE INDEX idx_purpose ON letters(sender,letter_purpose);`
	_, err = d.db.Exec(sqlStmt)
	if err != nil {
		err = errors.Wrap(err, "MakeTables, letters")
		return
	}

	return
}

// AddTag will add a tag into the database if it hasn't already been inserted.
func (d *database) AddTag(tag, id string) (err error) {
	if exists := d.TagExists(tag, id); exists {
		return
	}
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "AddTag")
	}
	stmt, err := tx.Prepare("INSERT INTO tags(tag,e_id) values (?, ?)")
	if err != nil {
		return errors.Wrap(err, "AddTag")
	}
	defer stmt.Close()

	_, err = stmt.Exec(tag, id)
	if err != nil {
		return errors.Wrap(err, "AddTag")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "AddTag")
	}

	return
}

// TagExists will return an error if the tag and ID does not exist
func (d *database) TagExists(tag, id string) (yes bool) {
	stmt, err := d.db.Prepare("SELECT e_id FROM tags WHERE tag = ? AND e_id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	var result string
	err = stmt.QueryRow(tag, id).Scan(&result)
	if err != nil {
		return
	}
	if result == "" {
		return
	}
	return true
}

// GetIDsFromTag get all the ides for a tag
func (d *database) GetIDsFromTag(tag string) (ids []string, err error) {
	stmt, err := d.db.Prepare("SELECT e_id FROM tags WHERE tag = ?")
	if err != nil {
		return nil, errors.Wrap(err, "problem preparing SQL")
	}
	defer stmt.Close()
	rows, err := stmt.Query(tag)
	if err != nil {
		return nil, errors.Wrap(err, "problem getting key")
	}
	ids = []string{}
	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			err = errors.Wrap(err, "GetIDsFromTag")
			return
		}
		ids = append(ids, id)
	}
	return
}

// Get will retrieve the value associated with a key.
func (d *database) Get(bucket, key string, v interface{}) (err error) {
	stmt, err := d.db.Prepare("select value from keystore where bucket_key = ?")
	if err != nil {
		return errors.Wrap(err, "problem preparing SQL")
	}
	defer stmt.Close()
	var result string
	err = stmt.QueryRow(bucket + "/" + key).Scan(&result)
	if err != nil {
		return errors.Wrap(err, "problem getting key")
	}

	err = json.Unmarshal([]byte(result), &v)
	if err != nil {
		return
	}
	return
}

// Set will set a value in the database, when using it like a keystore.
func (d *database) Set(bucket, key string, value interface{}) (err error) {
	var b []byte
	b, err = json.Marshal(value)
	if err != nil {
		return err
	}
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "Set")
	}
	stmt, err := tx.Prepare("insert or replace into keystore(bucket_key,value) values (?, ?)")
	if err != nil {
		return errors.Wrap(err, "Set")
	}
	defer stmt.Close()

	_, err = stmt.Exec(bucket+"/"+key, string(b))
	if err != nil {
		return errors.Wrap(err, "Set")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "Set")
	}

	return
}

// addEnvelope will add or replace an envelope
func (d *database) addEnvelope(e letter.Envelope) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return
	}
	var opened int
	// marshaled things
	var mSealedRecipients, mTo string
	if e.Opened {
		opened = 1
	} else {
		opened = 0
	}
	var b []byte

	b, err = json.Marshal(e.SealedRecipients)
	if err != nil {
		return errors.Wrap(err, "problem marshaling SealedRecipients")
	}
	mSealedRecipients = string(b)

	b, err = json.Marshal(e.Letter.To)
	if err != nil {
		return errors.Wrap(err, "problem marshaling To")
	}
	mTo = string(b)

	stmt, err := tx.Prepare("insert or replace into letters(id,time,sender,signature,sealed_recipients,sealed_letter,opened,letter_purpose,letter_to,letter_content,letter_firstid,letter_replyto) values(?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return
	}
	defer stmt.Close()
	_, err = stmt.Exec(e.ID, e.Timestamp, e.Sender.Public, e.Signature, mSealedRecipients, e.SealedLetter, opened, e.Letter.Purpose, mTo, e.Letter.Content, e.Letter.FirstID, e.Letter.ReplyTo)
	if err != nil {
		return
	}
	tx.Commit()
	return
}

func (d *database) getAllFromQuery(query string) (s []letter.Envelope, err error) {
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getAllFromQuery")
		return
	}
	defer rows.Close()

	// parse rows
	s, err = d.getRows(rows)
	if err != nil {
		err = errors.Wrap(err, query)
	}
	return
}

// getAllFromPreparedQuery
func (d *database) getAllFromPreparedQuery(query string, args ...interface{}) (s []letter.Envelope, err error) {
	// prepare statement
	logger.Log.Debug(query)
	stmt, err := d.db.Prepare(query)
	if err != nil {
		err = errors.Wrap(err, query)
		return
	}
	defer stmt.Close()
	rows, err := stmt.Query(args...)
	if err != nil {
		err = errors.Wrap(err, query)
		return
	}
	defer rows.Close()
	s, err = d.getRows(rows)
	if err != nil {
		err = errors.Wrap(err, query)
	}
	return
}

func (d *database) getRows(rows *sql.Rows) (s []letter.Envelope, err error) {
	// loop through rows
	s = []letter.Envelope{}
	for rows.Next() {
		var e letter.Envelope
		e.Letter = letter.Letter{}
		var opened int
		// marshaled things
		var mSender, mSealedRecipients, mTo string
		err = rows.Scan(&e.ID, &e.Timestamp, &mSender, &e.Signature, &mSealedRecipients, &e.SealedLetter, &opened, &e.Letter.Purpose, &mTo, &e.Letter.Content, &e.Letter.FirstID, &e.Letter.ReplyTo)
		e.Sender, err = keypair.FromPublic(mSender)
		json.Unmarshal([]byte(mSealedRecipients), &e.SealedRecipients)
		json.Unmarshal([]byte(mTo), &e.Letter.To)

		e.Opened = opened == 1
		if err != nil {
			err = errors.Wrap(err, "getRows")
			return
		}

		s = append(s, e)
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getRows")
	}
	return
}

// getKeys returns all the keys shared with you in the database, which can be queried by the sender
func (d *database) getKeys(sender ...string) (s []keypair.KeyPair, err error) {
	s = []keypair.KeyPair{}

	var query string
	if len(sender) > 0 {
		query = fmt.Sprintf("SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == '%s' AND sender == '%s' ORDER BY time DESC;", purpose.ShareKey, sender[0])
	} else {
		query = fmt.Sprintf("SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == '%s' ORDER BY time DESC;", purpose.ShareKey)
	}
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getKeys")
		return
	}
	defer rows.Close()

	// loop through rows
	for rows.Next() {
		var mKeyPair string
		err = rows.Scan(&mKeyPair)
		if err != nil {
			err = errors.Wrap(err, "getKeys")
			return
		}

		var kp keypair.KeyPair
		err = json.Unmarshal([]byte(mKeyPair), &kp)
		if err != nil {
			return
		}
		s = append(s, kp)
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getKeys")
	}
	return
}

// getIDs returns all the envelope IDs
func (d *database) getIDs(sender ...string) (s []string, err error) {
	s = []string{}
	query := fmt.Sprintf("SELECT id FROM letters ORDER BY time DESC;")
	if len(sender) > 0 {
		query = fmt.Sprintf("SELECT id FROM letters WHERE sender == '%s' ORDER BY time DESC;", sender[0])
	}
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getIDs")
		return
	}
	defer rows.Close()

	// loop through rows
	for rows.Next() {
		var mID string
		err = rows.Scan(&mID)
		if err != nil {
			err = errors.Wrap(err, "getIDs")
			return
		}
		s = append(s, mID)
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getIDs")
	}
	return
}

// getName returns the name of a person
func (d *database) getName(person string) (name string, err error) {
	query := fmt.Sprintf("SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == '%s' AND sender == '%s' ORDER BY time DESC;", purpose.ActionName, person)
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getName")
		return
	}
	defer rows.Close()

	// loop through rows
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			err = errors.Wrap(err, "getName")
			return
		}
		break
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getName")
	}
	return
}

// getProfile returns the profile of a person
func (d *database) getProfile(person string) (profile string, err error) {
	query := fmt.Sprintf("SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == '%s' AND sender == '%s' ORDER BY time DESC;", purpose.ActionProfile, person)
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getProfile")
		return
	}
	defer rows.Close()

	// loop through rows
	for rows.Next() {
		err = rows.Scan(&profile)
		if err != nil {
			err = errors.Wrap(err, "getProfile")
			return
		}
		break
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getProfile")
	}
	return
}

// getProfileImage returns the ID of the profile image of a person
func (d *database) getProfileImage(person string) (imageID string, err error) {
	query := fmt.Sprintf("SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == '%s' AND sender == '%s' ORDER BY time DESC;", purpose.ActionImage, person)
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getProfileImage")
		return
	}
	defer rows.Close()

	// loop through rows
	for rows.Next() {
		err = rows.Scan(&imageID)
		if err != nil {
			err = errors.Wrap(err, "getProfileImage")
			return
		}
		break
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getProfileImage")
	}
	return
}

func (d *database) getFriendsName(publicKey string) (name string) {
	query := "SELECT sender FROM letters WHERE opened == 1 AND letter_purpose == '" + purpose.ShareKey + "' AND letter_content LIKE '%" + publicKey + "%' LIMIT 1;"
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		return
	}
	defer rows.Close()

	var sender string
	for rows.Next() {
		rows.Scan(&sender)
	}
	if sender == "" {
		return
	}

	senderName, _ := d.getName(sender)
	if senderName != "" {
		sender = senderName
	}

	return "Friends of " + sender
}

// deleteLetterFromID will delete a letter with the pertaining ID.
func (d *database) deleteLetterFromID(id string) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteLetterFromID")
	}
	query := "DELETE FROM letters WHERE id == ?"
	logger.Log.Debug(query)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteLetterFromID")
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return errors.Wrap(err, "deleteLetterFromID")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteLetterFromID")
	}

	return
}

// deleteLettersFromSender will delete a letter with the pertaining ID.
func (d *database) deleteLettersFromSender(sender string) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteLettersFromSender")
	}
	query := "DELETE FROM letters WHERE sender == ?"
	logger.Log.Debug(query, sender)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteLettersFromSender")
	}
	defer stmt.Close()

	_, err = stmt.Exec(sender)
	if err != nil {
		return errors.Wrap(err, "deleteLettersFromSender")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteLettersFromSender")
	}

	return
}

// deleteUsersOldestPost will delete a letter with the pertaining ID.
func (d *database) deleteUsersOldestPost(publicKey string) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestPost")
	}
	logger.Log.Debug(publicKey)
	query := "DELETE from letters WHERE id in (SELECT id FROM letters WHERE letter_purpose IN ('share-text','share-image/png','share-image/jpg','') AND sender == ? ORDER BY time LIMIT 1);"
	logger.Log.Debug(query)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestPost")
	}
	defer stmt.Close()

	_, err = stmt.Exec(publicKey)
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestPost")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestPost")
	}
	return
}

// deleteUsersOldestPost will delete a letter with the pertaining ID.
func (d *database) deleteUsersOldestLargestPost(publicKey string) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestLargestPost")
	}
	logger.Log.Debug(publicKey)
	query := "DELETE from letters WHERE id in (SELECT id FROM letters WHERE LENGTH(sealed_letter) > 5000 AND sender == ? ORDER BY time LIMIT 1);"
	logger.Log.Debug(query)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestLargestPost")
	}
	defer stmt.Close()

	_, err = stmt.Exec(publicKey)
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestLargestPost")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldestLargestPost")
	}
	return
}

// deleteUser will delete everything except the ActionErases
func (d *database) deleteUsers() (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}
	query := "DELETE FROM letters WHERE sender IN (SELECT sender FROM letters WHERE letter_purpose == '" + purpose.ActionErase + "');"
	logger.Log.Debug(query)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}
	return
}

// deleteUser will delete everything except the ActionErases
func (d *database) deleteUser(publicKey string) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}
	query := "DELETE FROM letters WHERE sender == ?;"
	logger.Log.Debug(query)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}
	defer stmt.Close()

	_, err = stmt.Exec(publicKey)
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteUser")
	}
	return
}

// deleteUsersOldActions will delete old actions and leave only the most recent action undeleted
func (d *database) deleteUsersOldActions(publicKey string, purpose string) (err error) {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldActions")
	}
	logger.Log.Debug(publicKey, purpose)
	query := "DELETE FROM letters WHERE id in (SELECT id FROM letters WHERE opened == 1 AND letter_purpose == ? AND sender == ? ORDER BY time DESC LIMIT 1000000000 OFFSET 1);"
	logger.Log.Debug(query)
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldActions")
	}
	defer stmt.Close()

	_, err = stmt.Exec(purpose, publicKey)
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldActions")
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "deleteUsersOldActions")
	}
	return
}

// deleteUsersEdits will delete a letter with the pertaining ID.
func (d *database) deleteUsersEdits(publicKey string) (err error) {
	ids, err := d.getIDs(publicKey)
	if err != nil {
		return
	}
	for _, id := range ids {
		idVersions, err := d.getAllVersions(id)
		if err != nil {
			continue
		}
		for i, idVersion := range idVersions {
			if i == 0 {
				continue
			}
			d.deleteLetterFromID(idVersion)
		}
	}
	return
}

// isReplaced will see if there is more than one of a first_id, which only happens if something has been replaced.
func (d *database) isReplaced(id string) (yes bool, err error) {
	stmt, err := d.db.Prepare("SELECT COUNT(id) WHERE opened == 1 AND letter_firstid == ?")
	if err != nil {
		err = errors.Wrap(err, "problem preparing SQL")
		return
	}
	defer stmt.Close()
	var num int
	err = stmt.QueryRow(id).Scan(&num)
	if err != nil {
		return false, errors.Wrap(err, "problem getting")
	}
	yes = num > 1
	return
}

func (d *database) diskSpaceForUser(user string) (diskSpace int64, err error) {
	diskSpace = 0
	stmt, err := d.db.Prepare("SELECT SUM(LENGTH(sealed_letter))+SUM(LENGTH(sealed_recipients)) FROM letters WHERE sender==?")
	if err != nil {
		err = errors.Wrap(err, "problem preparing SQL")
		return
	}
	defer stmt.Close()
	stmt.QueryRow(user).Scan(&diskSpace)
	return
}

func (d *database) numLikesPerPost(idPost string) (likes int64, err error) {
	stmt, err := d.db.Prepare("SELECT COUNT(id) FROM letters WHERE opened == 1 AND letter_purpose == '" + purpose.ActionLike + "' AND letter_content=?")
	if err != nil {
		err = errors.Wrap(err, "problem preparing SQL")
		return
	}
	defer stmt.Close()
	err = stmt.QueryRow(idPost).Scan(&likes)
	if err != nil {
		err = errors.Wrap(err, "problem getting")
	}
	return
}

func (d *database) listUsers() (s []string, err error) {
	query := fmt.Sprintf("SELECT DISTINCT(sender) FROM letters;")
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "listUsers")
		return
	}
	defer rows.Close()

	s = []string{}
	for rows.Next() {
		var mID string
		err = rows.Scan(&mID)
		if err != nil {
			err = errors.Wrap(err, "listUsers")
			return
		}
		s = append(s, mID)
	}
	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "listUsers")
	}
	return
}

func (d *database) listBlockedUsers(publicKey string) (s []string, err error) {
	query := fmt.Sprintf("SELECT letter_content FROM letters WHERE opened == 1 AND letter_purpose == '%s' AND sender == '%s' AND letter_content != '';", purpose.ActionBlock, publicKey)
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "listBlockedUsers")
		return
	}
	defer rows.Close()

	s = []string{}
	for rows.Next() {
		var mID string
		err = rows.Scan(&mID)
		if err != nil {
			err = errors.Wrap(err, "listBlockedUsers")
			return
		}
		s = append(s, mID)
	}

	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "listBlockedUsers")
	}
	return
}

func (d *database) getFollowing(publicKey string) (s []string, err error) {
	query := fmt.Sprintf("SELECT DISTINCT(letter_content) FROM letters WHERE opened == 1 AND letter_purpose == '%s' AND sender == '%s' ;", purpose.ActionFollow, publicKey)
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getFollowing")
		return
	}
	defer rows.Close()

	// parse rows
	s = []string{}
	for rows.Next() {
		var mID string
		err = rows.Scan(&mID)
		if err != nil {
			err = errors.Wrap(err, "getFollowing")
			return
		}
		if mID == "" {
			continue
		}
		s = append(s, mID)
	}
	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getFollowing")
	}
	return
}

func (d *database) getFollowers(publicKey string) (s []string, err error) {
	query := fmt.Sprintf("SELECT DISTINCT(sender) FROM letters WHERE opened==1 AND letter_purpose == '%s' AND letter_content == '%s';", purpose.ActionFollow, publicKey)
	logger.Log.Debug(query)
	rows, err := d.db.Query(query)
	if err != nil {
		err = errors.Wrap(err, "getFollowers")
		return
	}
	defer rows.Close()

	s = []string{}
	for rows.Next() {
		var mID string
		err = rows.Scan(&mID)
		if err != nil {
			err = errors.Wrap(err, "getFollowers")
			return
		}
		if mID == "" {
			continue
		}
		s = append(s, mID)
	}
	err = rows.Err()
	if err != nil {
		err = errors.Wrap(err, "getFollowers")
	}
	return
}

func (d *database) getAllVersions(id string) (s []string, err error) {
	es, err := d.getAllFromPreparedQuery("SELECT * FROM letters WHERE opened == 1 AND letter_firstid == (SELECT letter_firstid FROM letters WHERE opened ==1 AND id == ?) ORDER BY time DESC", id)
	if err != nil {
		return
	}
	if len(es) == 0 {
		err = errors.New("no letters")
		return
	}

	s = make([]string, len(es))
	for i, e := range es {
		s[i] = e.ID
	}
	return
}

func (d *database) getKeyForFriends(user string) (key keypair.KeyPair, err error) {
	stmt, err := d.db.Prepare("SELECT letter_content FROM letters WHERE opened ==1 AND letter_purpose==? AND sender==? ORDER BY time DESC")
	if err != nil {
		err = errors.Wrap(err, "getKeyForFriends, bad statement")
		return
	}
	defer stmt.Close()
	var keystring string
	err = stmt.QueryRow(purpose.ShareKey, user).Scan(&keystring)
	if err != nil {
		err = errors.Wrap(err, "getKeyForFriends, bad scan")
		return
	}
	logger.Log.Infof(`query: "SELECT letter_content FROM letters WHERE opened ==1 AND letter_purpose=='%s' AND sender=='%s' ORDER BY time DESC": [%v]`, purpose.ShareKey, user, keystring)
	err = json.Unmarshal([]byte(keystring), &key)
	if err != nil {
		err = errors.Wrap(err, "getKeyForFriends, bad unmarshal")
		return
	}

	return
}
