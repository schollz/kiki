package feed

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	strip "github.com/schollz/html-strip-tags-go"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/purpose"
	"github.com/schollz/kiki/src/utils"
	"github.com/schollz/kiki/src/web"
)

// New generates a new feed based on the location to find the identity file, the database, and the settings
func New(location ...string) (f Feed, err error) {
	locationToSaveData := "."
	if len(location) > 0 {
		locationToSaveData = location[0]
	}

	locationToSaveData, err = filepath.Abs(locationToSaveData)
	if err != nil {
		return
	}
	f = Feed{
		storagePath: locationToSaveData,
		Settings:    GenerateSettings(),
	}

	// initialize
	err = f.init()
	return
}

// Open will load a feed from the specified location
func Open(locationToFeed string) (f Feed, err error) {
	bFeed, err := ioutil.ReadFile(path.Join(locationToFeed, "feed.json"))
	if err != nil {
		return
	}
	err = json.Unmarshal(bFeed, &f)
	if err != nil {
		return
	}
	f.storagePath = locationToFeed

	// initialize
	err = f.init()
	return
}

// init initializes the kiki instance
func (f *Feed) init() (err error) {
	f.log = logging.Log
	f.log.Debug("initializing feed")
	loc, _ := filepath.Abs(f.storagePath)
	f.log.Infof("database location: %s", loc)

	if f.RegionKey.Public == "" {
		// define region key
		f.RegionKey, err = keypair.FromPair("rbcDfDMIe8qXq4QPtIUtuEylDvlGynx56QgeHUZUZBk=",
			"GQf6ZbBbnVGhiHZ_IqRv0AlfqQh1iofmSyFOcp1ti8Q=") // define region key
		if err != nil {
			return
		}
	}

	f.db = database.Setup(f.storagePath)

	if f.PersonalKey.Public == "" {
		// generate a new personal key
		var err2 error
		f.PersonalKey = keypair.New()

		// add the friends key
		err2 = f.AddFriendsKey()
		if err2 != nil {
			return err2
		}

		// block the region public key from being used as a sender, ever
		err2 = f.ProcessLetter(letter.Letter{
			To:      []string{"public"},
			Purpose: purpose.ActionBlock,
			Content: f.RegionKey.Public,
		})
		if err2 != nil {
			err2 = errors.Wrap(err2, "setup")
			return err2
		}
	}

	// overwrite the feed file
	feedBytes, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path.Join(f.storagePath, "feed.json"), feedBytes, 0644)

	return
}

func (f Feed) Cleanup() {
	fmt.Println("cleaning up...")
}

func (f Feed) DoSyncing() {
	for {
		for _, server := range f.Settings.AvailableServers {
			f.Sync(server)
		}
		time.Sleep(60 * time.Second)
	}
}

// ProcessLetter will determine where to put the letter
func (f Feed) ProcessLetter(l letter.Letter) (err error) {
	if !purpose.Valid(l.Purpose) {
		err = errors.New("invalid purpose")
		return
	}
	if f.PersonalKey == f.RegionKey {
		err = errors.New("cannot post with region key")
		return
	}
	if l.Replaces != "" {
		e, err2 := f.db.GetEnvelopeFromID(l.Replaces)
		if err2 != nil {
			return errors.New("problem replacing that")
		}
		if f.PersonalKey.Public != e.Sender.Public {
			return errors.New("refusing to replace someone else's post")
		}
	}

	if strings.Contains(l.Purpose, "action-") {
		// assignments are always public
		l.To = []string{f.RegionKey.Public}
	} else {
		// rewrite the letter.To array so that it contains
		// public keys that are valid
		newTo := []string{}
		for _, to := range l.To {
			switch to {
			case "public":
				newTo = append(newTo, f.RegionKey.Public)
			case "self":
				// automatically done when adding any letter
				// this just put here for pedantic reasons
			case "friends":
				friendsKeyPairs, err2 := f.db.GetKeysFromSender(f.PersonalKey.Public)
				if err2 != nil {
					return err2
				}
				for _, friendsKeyPair := range friendsKeyPairs {
					newTo = append(newTo, friendsKeyPair.Public)
				}
			default:
				_, err2 := keypair.FromPublic(to)
				if err2 != nil {
					f.log.Infof("Not a valid public key: '%s'", to)
				} else {
					newTo = append(newTo, to)
				}
			}
		}
		l.To = newTo
	}

	// determine if their are any images in envelope letter content that should be spliced out
	newHTML, images, err := web.CaptureBase64Images(l.Content)
	if err != nil {
		return
	}
	f.log.Debugf("found %d images", len(images))
	for name := range images {
		p := purpose.SharePNG
		if strings.Contains(name, ".jpg") {
			p = purpose.ShareJPG
		}
		newLetter := letter.Letter{
			To:      l.To,
			Content: base64.URLEncoding.EncodeToString(images[name]),
			Purpose: p,
		}
		newEnvelope, err2 := newLetter.Seal(f.PersonalKey, f.RegionKey)
		if err2 != nil {
			return err2
		}
		// seal and add envelope
		err2 = f.db.AddEnvelope(newEnvelope)
		if err2 != nil {
			return err2
		}
		if l.Purpose == purpose.ActionImage {
			newHTML = newEnvelope.ID
			break
		}
		newHTML = strings.Replace(newHTML, name, newEnvelope.ID, 1)
	}
	l.Content = newHTML

	// remove tags from name change
	if l.Purpose == purpose.ActionName {
		l.Content = strip.StripTags(l.Content)
	}
	if strip.StripTags(l.Content) == "" {
		l.Content = ""
	}

	// seal the letter
	e, err := l.Seal(f.PersonalKey, f.RegionKey)
	if err != nil {
		return
	}
	err = f.db.AddEnvelope(e)
	if err != nil {
		return
	}

	err = f.UnsealLetters()
	return
}

// ProcessEnvelope will determine whether the incoming letter is valid and can be submitted to the database.
func (f Feed) ProcessEnvelope(e letter.Envelope) (err error) {
	// check if envelope has a valid signature
	err = e.Validate(f.RegionKey)
	if err != nil {
		return
	}

	// check if envelope already exists
	_, errGet := f.GetEnvelope(e.ID)
	if errGet == nil {
		return errors.New("already have envelope")
	}

	err = f.db.AddEnvelope(e)
	if err != nil {
		return
	}

	err = f.UnsealLetters()
	return
}

// UnsealLetters will go through unopened envelopes and open them and then add them to the f.db. Also go through and purge bad letters (invalidated letters)
func (f Feed) UnsealLetters() (err error) {
	lettersToPurge := []string{}
	envelopes, err := f.db.GetAllEnvelopes(false)
	if err != nil {
		return err
	}

	// get friends keys
	keysToTry, err := f.db.GetKeys()
	if err != nil {
		return
	}
	// prepend public key
	keysToTry = append([]keypair.KeyPair{f.RegionKey}, keysToTry...)
	// add personal key last
	keysToTry = append(keysToTry, f.PersonalKey)
	for _, envelope := range envelopes {
		if err := envelope.Validate(f.RegionKey); err != nil {
			// add to purge
			lettersToPurge = append(lettersToPurge, envelope.ID)
		}
		ue, err := envelope.Unseal(keysToTry, f.RegionKey)
		if err != nil {
			f.log.Debug(err)
			continue
		}
		f.log.Debug(ue.Letter)
		err = f.db.AddEnvelope(ue)
		if err != nil {
			f.log.Debug(err)
			continue
		}
	}

	// purge invalid letters
	if len(lettersToPurge) > 0 {
		err = f.db.RemoveLetters(lettersToPurge)
	}
	return
}

func (f Feed) ShowProfile() (u User, err error) {
	name, profile, image := f.db.GetUser(f.PersonalKey.Public)
	u = User{
		Name:      strip.StripTags(name),
		PublicKey: f.PersonalKey.Public,
		Profile:   template.HTML(profile),
		Image:     image,
	}
	return
}

type ShowFeedParameters struct {
	ID      string // view a single post
	Channel string // filter by channel
	User    string // filter by user
	Search  string // filter by search term
	Latest  bool   // get the latest
}

func (f Feed) ShowFeed(p ShowFeedParameters) (posts []Post, err error) {
	var envelopes []letter.Envelope
	if p.ID != "" {
		envelopes = make([]letter.Envelope, 1)
		if p.Latest {

		} else {
			envelopes[0], err = f.db.GetEnvelopeFromID(p.ID)
		}
	} else if p.Channel != "" {

	} else if p.User != "" {

	} else if p.Search != "" {

	} else {
		// reteurn all envelopes
		envelopes, err = f.db.GetBasicPosts()
	}
	f.log.Debugf("Found %d envelopes", len(envelopes))
	posts = make([]Post, len(envelopes))
	i := 0
	for _, e := range envelopes {
		post := f.MakePost(e)
		comments := f.DetermineComments(post.ID)
		posts[i] = Post{
			Post:     post,
			Comments: comments,
		}
		i++
	}
	posts = posts[:i]
	fmt.Println(posts)
	return
}

func (f Feed) MakePost(e letter.Envelope) (post BasicPost) {
	recipients := []string{}
	for _, to := range e.Letter.To {
		if to == f.RegionKey.Public {
			recipients = []string{"Public"}
			break
		}
		friendsName := strip.StripTags(f.db.GetFriendsName(to))
		if friendsName != "" {
			recipients = []string{friendsName}
			break
		}
		senderName := f.db.GetName(to)
		if senderName == "" {
			senderName = to
		}
		recipients = append(recipients, senderName)
	}

	// convert UTC timestamps to local time
	timeLocation, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}
	convertedTime := e.Timestamp.In(timeLocation)

	post = BasicPost{
		ID:         e.ID,
		Recipients: strings.Join(recipients, ", "),
		Content:    template.HTML(e.Letter.Content),
		Date:       convertedTime,
		TimeAgo:    utils.TimeAgo(convertedTime),
		User: User{
			Name:      strip.StripTags(f.db.GetName(e.Sender.Public)),
			PublicKey: e.Sender.Public,
			Profile:   template.HTML(f.db.GetProfile(e.Sender.Public)),
			Image:     f.db.GetProfileImage(e.Sender.Public),
		},
	}
	return
}

func (f Feed) DetermineComments(postID string) []BasicPost {
	return f.recurseComments(postID, []BasicPost{}, 0)
}

func (f Feed) recurseComments(postID string, comments []BasicPost, depth int) []BasicPost {
	es, err := f.db.GetReplies(postID)
	if err != nil {
		f.log.Error(err)
	}
	for _, e := range es {
		comment := f.MakePost(e)
		comment.Depth = depth
		comment.ReplyTo = postID
		comments = append(comments, comment)
		comments = f.recurseComments(comment.ID, comments, depth+1)
	}
	return comments
}

// AddFriendsKey will generate a new friends key and post it to the feed
func (f Feed) AddFriendsKey() (err error) {
	// generate a key for friends
	myfriends := keypair.New()
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey")
		return
	}
	myfriendsByte, err := json.Marshal(myfriends)

	// share the friends key with yourself
	err = f.ProcessLetter(letter.Letter{
		To:      []string{"self"},
		Purpose: purpose.ShareKey,
		Content: string(myfriendsByte),
	})
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey")
		return
	}

	// block the friends public key from being used as a sender, ever
	err = f.ProcessLetter(letter.Letter{
		To:      []string{"public"},
		Purpose: purpose.ActionBlock,
		Content: myfriends.Public,
	})
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey")
		return
	}

	return
}

// GetEnvelope will return an envelope with the given ID
func (f Feed) GetEnvelope(id string) (e letter.Envelope, err error) {
	return f.db.GetEnvelopeFromID(id)
}

// GetIDs will return an envelope with the given ID
func (f Feed) GetIDs() (ids map[string]struct{}, err error) {
	return f.db.GetIDs()
}

// Sync will try to sync with the respective address
func (f Feed) Sync(address string) (err error) {
	// make sure that its a kiki instance
	isInstance, err := f.IsKikiInstance(address)
	if err != nil {
		return
	}
	if !isInstance {
		return errors.New("not a kiki instance")
	}

	f.log.Debugf("syncing with %s", address)

	// Get a list of my IDs
	myIDs, err := f.GetIDs()
	if err != nil {
		return
	}

	var target Response
	f.log.Debug("getting list")
	req, err := http.NewRequest("GET", address+"/list", nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	f.log.Debug("unmarshaling response")
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return
	}
	f.log.Debug(target)
	if "ok" != target.Status {
		return errors.New(target.Error)
	}
	if target.RegionPublicKey != f.RegionKey.Public {
		return errors.New("cannot sync with another region")
	}

	f.log.Debugf("got %d IDs from %s", len(target.IDs), address)

	// check whether I need any of their envelopes
	for theirID := range target.IDs {
		if _, ok := myIDs[theirID]; ok {
			continue
		}
		f.log.Debugf("%s has new envelope: %s", address, theirID)
		err = f.DownloadEnvelope(address, theirID)
		if err != nil {
			return
		}
	}

	// check whether they need any of my envelopes
	for myID := range myIDs {
		if _, ok := target.IDs[myID]; ok {
			continue
		}
		f.log.Debugf("my envelope %s is new to %s", myID, address)
		err = f.UploadEnvelope(address, myID)
		if err != nil {
			return
		}
	}
	return
}

// UploadEnvelope will upload the specified envelope
func (f Feed) UploadEnvelope(address, id string) (err error) {
	// get envelope
	e, err := f.GetEnvelope(id)
	if err != nil {
		return
	}
	// close it
	e.Close()

	// marshal it
	payloadBytes, err := json.Marshal(e)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	// POST it
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/envelope", address), body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var target Response
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return
	}
	if "ok" != target.Status {
		return errors.New(target.Error)
	}

	f.log.Debugf("uploaded %s to %s", id, address)

	return
}

// DownloadEnvelope will download the specified envelope
func (f Feed) DownloadEnvelope(address, id string) (err error) {
	req, err := http.NewRequest("GET", address+"/download/"+id, nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var target Response
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return
	}
	if "ok" != target.Status {
		return errors.New(target.Error)
	}

	f.log.Debugf("downloaded %s from %s", target.Envelope.ID, address)

	err = f.ProcessEnvelope(target.Envelope)
	return
}

// IsKikiInstance will download the specified envelope
func (f Feed) IsKikiInstance(address string) (yes bool, err error) {
	timeout := time.Duration(1500 * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Get(address + "/ping")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var target Response
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return
	}
	if "ok" != target.Status {
		err = errors.New(target.Error)
		return
	}

	if target.Message == f.RegionKey.Public {
		yes = true
	}
	return
}

// PurgeOverflowingStorage will delete old messages
func (f Feed) PurgeOverflowingStorage() (err error) {
	f.log.Debug(f.db.ListUsers())
	f.log.Debug(f.db.DiskSpaceForUser(f.PersonalKey.Public))
	users, err := f.db.ListUsers()
	if err != nil {
		return
	}
	for _, user := range users {
		// skip personal user
		if user == f.PersonalKey.Public {
			continue
		}
		// skip friends
		// TODO
		f.log.Debug(user)
		f.log.Debug(f.db.DiskSpaceForUser(user))
	}
	return
}

func (f Feed) TestStuff() {
	f.log.Debug(f.db.GetAllVersions("9f94a236940326a7fb9eb06d3cc0fece47dee6fa58b8d328b47a9ebc175d7616"))
}
