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
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"

	"github.com/pkg/errors"
	cache "github.com/robfig/go-cache"
	strip "github.com/schollz/html-strip-tags-go"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/purpose"
	"github.com/schollz/kiki/src/utils"
	"github.com/schollz/kiki/src/web"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func (f *Feed) Debug(b bool) {
	if !b {
		f.logger.SetLevel("warn")
	} else {
		f.logger.SetLevel("debug")
	}
	f.log = logging.Log
	database.Debug(b)
}

// New generates a new feed based on the location to find the identity file, the database, and the settings
func New(params ...string) (f *Feed, err error) {
	regionKeyPublic := "rbcDfDMIe8qXq4QPtIUtuEylDvlGynx56QgeHUZUZBk="
	regionKeyPrivate := "GQf6ZbBbnVGhiHZ_IqRv0AlfqQh1iofmSyFOcp1ti8Q="
	locationToSaveData := "."
	if len(params) > 0 {
		locationToSaveData = params[0]
		if len(params) == 3 {
			regionKeyPublic = params[1]
			regionKeyPrivate = params[2]
		}
	}

	locationToSaveData, err = filepath.Abs(locationToSaveData)
	if err != nil {
		return
	}

	f = new(Feed)
	f.Settings = GenerateSettings()
	f.db = database.Setup(locationToSaveData)
	f.storagePath = locationToSaveData
	f.logger = logging.New()
	f.caching = cache.New(1*time.Minute, 5*time.Minute)
	f.servers.Lock()
	f.servers.connected = make(map[string]User)
	f.servers.blockedUsers = make(map[string]struct{})
	f.servers.syncingCount = 0
	f.servers.Unlock()
	f.logger.Log.Infof("feed located at: '%s'", f.storagePath)
	bFeed, errLoad := ioutil.ReadFile(path.Join(f.storagePath, "kiki.json"))
	if errLoad != nil {
		fmt.Println("generating new feed")

		// define region key
		err = f.SetRegionKey(regionKeyPublic,
			regionKeyPrivate)
		if err != nil {
			return
		}

		// generate a new personal key
		var err2 error
		f.PersonalKey = keypair.New()

		// add the friends key
		err2 = f.AddFriendsKey()
		if err2 != nil {
			err = errors.Wrap(err2, "add the friends key")
			return
		}

		if err2 != nil {
			err = errors.Wrap(err2, "setup")
			return
		}

		// send welcome messasge
		err2 = f.ProcessLetter(letter.Letter{
			To:      []string{},
			Purpose: purpose.ShareText,
			Content: `<p>Welcome to KiKi!</p><p>To get started, you can change your name, edit your profile, upload an image, and make posts!</p> `,
		})
		if err2 != nil {
			err = errors.Wrap(err2, "setup")
			return
		}

	} else {
		err = json.Unmarshal(bFeed, &f)
		if err != nil {
			return
		}
	}

	err = f.Save()
	f.UpdateEverything()

	go f.UpdateOnUpload()
	return
}

func (f *Feed) SetRegionKey(public, private string) (err error) {
	f.RegionKey, err = keypair.FromPair(public, private) // define region key
	if err != nil {
		return
	}
	return
}

func (f *Feed) Save() (err error) {
	// overwrite the feed file
	feedBytes, err := json.MarshalIndent(f, "", " ")
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path.Join(f.storagePath, "kiki.json"), feedBytes, 0644)
	return
}

func (f *Feed) Cleanup() {
	fmt.Println("cleaning up...")
}

func (f *Feed) UpdateBlockedUsers() (err error) {
	// update the blocked users
	blockedUsers, err := f.db.ListBlockedUsers(f.PersonalKey.Public)
	if err != nil {
		return
	}
	f.servers.Lock()
	f.servers.blockedUsers = make(map[string]struct{})
	for _, blockedUser := range blockedUsers {
		f.servers.blockedUsers[blockedUser] = struct{}{}
		f.db.RemoveLettersForUser(blockedUser)
	}
	f.servers.Unlock()
	return
}

func (f *Feed) SignalUpdate() {
	f.logger.Log.Info("signaling")
	f.servers.Lock()
	f.servers.syncingCount++
	f.servers.Unlock()
}

func (f *Feed) UpdateOnUpload() {
	for {
		time.Sleep(1 * time.Second)
		currentCount := 0
		f.servers.RLock()
		currentCount = f.servers.syncingCount
		f.servers.RUnlock()
		if currentCount > 0 {
			f.logger.Log.Debug("going to try to sync!")
			// wait three seconds and see if we have the same current count
			time.Sleep(3 * time.Second)
			f.servers.RLock()
			currentCount2 := f.servers.syncingCount
			f.servers.RUnlock()
			if currentCount != currentCount2 {
				continue
			}
			// if the count is stabilized, then do syncing
			f.UpdateEverything()
			f.servers.Lock()
			f.servers.syncingCount = 0
			f.servers.Unlock()
		}
	}
}

func (f *Feed) UpdateEverythingAndSync() {
	f.servers.Lock()
	if f.servers.updating {
		f.servers.Unlock()
		return
	}
	f.servers.updating = true
	f.servers.Unlock()

	f.UpdateEverything()
	f.SyncServers()

	f.servers.Lock()
	f.servers.updating = false
	f.servers.Unlock()
}

func (f *Feed) UpdateEverything() {
	f.logger.Log.Info("updating everything")
	// unseal any new letters
	err := f.UnsealLetters()
	if err != nil {
		f.logger.Log.Warn(err)
	}

	// send out friends keys for new friends
	err = f.UpdateFriends()
	if err != nil {
		f.logger.Log.Warn(err)
	}

	// update blocked users
	err = f.UpdateBlockedUsers()
	if err != nil {
		f.logger.Log.Warn(err)
	}

	// purge overflowing storage
	err = f.PurgeOverflowingStorage()
	if err != nil {
		f.logger.Log.Warn(err)
	}

	// erase profiles that want to be deleted
	err = f.db.DeleteProfiles()
	if err != nil {
		f.logger.Log.Error(err)
	}

	// erase things that are posted as shared keys or as region key
	f.db.DeleteUser(f.RegionKey.Public)
	keys, _ := f.db.GetKeys()
	for _, key := range keys {
		err2 := f.db.DeleteUser(key.Public)
		if err2 != nil {
			f.logger.Log.Error(err)
		}
	}

	// determine the available hashtags
	err = f.DetermineHashtags()
	if err != nil {
		f.logger.Log.Error(err)
	}
}

// DetermineHashtags will go through and find all the hashtags
func (f *Feed) DetermineHashtags() (err error) {
	r, err := regexp.Compile(`(\#[a-z-A-Z]+\b)`)
	if err != nil {
		return
	}
	es, err := f.db.GetAllEnvelopes(true)
	if err != nil {
		return
	}
	tagCounts := make(map[string]int)
	idToTags := make(map[string][]string)
	for _, e := range es {
		foundTags := make(map[string]struct{})
		idToTags[e.ID] = []string{}
		for _, tag := range r.FindAll([]byte(e.Letter.Content), -1) {
			t := strings.ToLower(string(tag))
			if len(t) < 3 {
				continue
			}

			foundTags[t[1:len(t)]] = struct{}{}
			idToTags[e.ID] = append(idToTags[e.ID], t[1:len(t)])
		}
		for tag := range foundTags {
			if _, ok := tagCounts[tag]; !ok {
				tagCounts[tag] = 0
			}
			tagCounts[tag]++
		}
	}
	f.logger.Log.Debugf("Found %d tags", len(tagCounts))
	f.logger.Log.Info(tagCounts)
	err = f.db.Set("globals", "tags", tagCounts)
	if err != nil {
		f.logger.Log.Error(err)
	}
	err = f.db.AddTags(idToTags)
	return
}

func (f *Feed) GetHashTags() (tags []string) {
	tagCounts := make(map[string]int)
	err := f.db.Get("globals", "tags", &tagCounts)
	if err != nil {
		f.logger.Log.Error(err)
		return
	}
	tags = make([]string, len(tagCounts))
	// TODO: Sort tags by popularity? or alphabetically?
	i := 0
	for tag := range tagCounts {
		tags[i] = tag
		i++
	}
	return
}

func (f *Feed) SyncServers() {
	f.logger.Log.Info("Starting syncing")
	needToUpdate := false
	for _, server := range f.Settings.AvailableServers {
		err := f.Sync(server)
		if err != nil {
			f.logger.Log.Warn(err)
		} else {
			needToUpdate = true
		}
	}
	if needToUpdate {
		f.UpdateEverything()
	}
}

// ProcessLetter will determine where to put the letter
func (f *Feed) ProcessLetter(l letter.Letter) (err error) {
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
		if e.Letter.ReplyTo != "" {
			l.ReplyTo = e.Letter.ReplyTo
		}
	}

	if strings.Contains(l.Purpose, "action-") {
		// actions are always public
		l.To = []string{f.RegionKey.Public}
		if l.Purpose == purpose.ActionBlock && l.Content == f.PersonalKey.Public {
			return errors.New("refusing to block yourself")
		} else if l.Purpose == purpose.ActionFollow && l.Content == f.PersonalKey.Public {
			return errors.New("refusing to follow yourself")
		}
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
				alreadyAdded := make(map[string]struct{})
				for _, friendsKeyPair := range friendsKeyPairs {
					if _, ok := alreadyAdded[friendsKeyPair.Public]; ok {
						continue
					}
					newTo = append(newTo, friendsKeyPair.Public)
					alreadyAdded[friendsKeyPair.Public] = struct{}{}
				}
			default:
				_, err2 := keypair.FromPublic(to)
				if err2 != nil {
					f.logger.Log.Infof("Not a valid public key: '%s'", to)
				} else if to == f.RegionKey.Public {
					f.logger.Log.Info("cannot post as public!")
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
	if l.Purpose == purpose.ShareText {
		// convert Markdown -> HTML
		l.Content = string(blackfriday.Run([]byte(l.Content)))
		// sanitize
		p := bluemonday.UGCPolicy()
		l.Content = p.Sanitize(l.Content)
		// replace hashtags with links to the hash tags
		r, _ := regexp.Compile(`(\#[a-z-A-Z]+\b)`)
		tags := r.FindAllString(l.Content, -1)
		tagMap := make(map[string]struct{})
		for _, tag := range tags {
			tagMap[tag] = struct{}{}
		}
		for tag := range tagMap {
			l.Content = strings.Replace(l.Content, tag, fmt.Sprintf(`<a href="/?hashtag=%s" class="hashtag">%s</a>`, tag, tag), -1)
		}
	}
	l.Content = strings.TrimSpace(l.Content)

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

	return
}

// ProcessEnvelope will determine whether the incoming letter is valid and can be submitted to the database.
func (f *Feed) ProcessEnvelope(e letter.Envelope) (err error) {
	// check if envelope has a valid signature
	err = e.Validate(f.RegionKey)
	if err != nil {
		return errors.Wrap(err, "ProcessEnvelope, not validated")
	}

	// check if envelope comes from blocked user
	f.servers.RLock()
	if _, ok := f.servers.blockedUsers[e.Sender.Public]; ok {
		f.servers.RUnlock()
		return errors.New("this user has been blocked, not downloading")
	}
	f.servers.RUnlock()

	// check if the storage limits are exceeded for this envelope
	// and then only accept if it is a newer envelope
	// TODO

	// check if envelope already exists
	_, errGet := f.GetEnvelope(e.ID)
	if errGet == nil {
		f.logger.Log.Debugf("skipping %s, already have", e.ID)
		// already have return
		return nil
	}

	err = f.db.AddEnvelope(e)
	if err != nil {
		return
	}

	return
}

// UnsealLetters will go through unopened envelopes and open them and then add them to the f.db. Also go through and purge bad letters (invalidated letters)
func (f *Feed) UnsealLetters() (err error) {
	lettersToPurge := []string{}
	envelopes, err := f.db.GetAllEnvelopes(false)
	if err != nil {
		return err
	}

	// get friends keys
	keysToTry, err := f.db.GetKeys()
	if err != nil {
		err = errors.Wrap(err, "UnsealLetters, getting keys")
		return
	}
	f.logger.Log.Debugf("Have %d keys from friends", len(keysToTry))
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
			// this user is not a recipient, just continue
			continue
		}
		err = f.db.UpdateEnvelope(ue)
		if err != nil {
			continue
		}
	}

	// purge invalid letters
	if len(lettersToPurge) > 0 {
		err = f.db.RemoveLetters(lettersToPurge)
	}
	return
}

// GetUser returns the information for a specific user
func (f *Feed) GetUser(public ...string) (u User) {
	publicKey := f.PersonalKey.Public
	if len(public) > 0 {
		publicKey = public[0]
	}
	name, profile, image := f.db.GetUser(publicKey)
	followers, following, friends := f.db.Friends(publicKey)
	blocked, _ := f.db.ListBlockedUsers(publicKey)
	u = User{
		Name:      strip.StripTags(name),
		PublicKey: publicKey,
		Profile:   template.HTML(profile),
		Image:     image,
		Followers: followers,
		Following: following,
		Friends:   friends,
		Blocked:   blocked,
	}
	return
}

// GetUserFriends returns detailed friend information
func (f *Feed) GetUserFriends() (u UserFriends) {
	followers, following, friends := f.db.Friends(f.PersonalKey.Public)
	u.Followers = make([]User, len(followers))
	for i := range followers {
		u.Followers[i] = f.GetUser(followers[i])
	}
	u.Following = make([]User, len(following))
	for i := range following {
		u.Following[i] = f.GetUser(following[i])
	}
	u.Friends = make([]User, len(friends))
	for i := range friends {
		u.Friends[i] = f.GetUser(friends[i])
	}
	return
}

// UpdateFriends will post keys to friends
func (f *Feed) UpdateFriends() (err error) {
	friendsKey, err := f.db.GetLatestKeyForFriends(f.PersonalKey.Public)
	if err != nil {
		err = errors.Wrap(err, "UpdateFriends")
		return
	}
	bFriendsKey, err := json.Marshal(friendsKey)
	if err != nil {
		err = errors.Wrap(err, "UpdateFriends")
		return
	}
	_, _, friends := f.db.Friends(f.PersonalKey.Public)
	for _, friend := range friends {
		l := letter.Letter{
			To:      []string{friend},
			Purpose: purpose.ShareKey,
			Content: string(bFriendsKey),
		}
		f.logger.Log.Debugf("Adding letter for friend %s", friend)
		f.ProcessLetter(l)
	}
	return
}

type ShowFeedParameters struct {
	ID      string // view a single post
	Hashtag string // filter by channel
	User    string // filter by user
	Search  string // filter by search term
	Latest  bool   // get the latest
}

func (f *Feed) ShowFeed(p ShowFeedParameters) (posts []Post, err error) {
	t := time.Now()
	var envelopes []letter.Envelope
	if p.ID != "" {
		envelopes = make([]letter.Envelope, 1)
		if p.Latest {
			envelopes[0], err = f.db.GetLatestEnvelopeFromID(p.ID)
		} else {
			envelopes[0], err = f.db.GetEnvelopeFromID(p.ID)
		}
	} else if p.Hashtag != "" {
		envelopes, err = f.db.GetEnvelopesFromTag(strings.ToLower(p.Hashtag))
		f.logger.Log.Debugf("Got %d envelopes searching for '#%s'", len(envelopes), p.Hashtag)
	} else if p.User != "" {
		f.logger.Log.Debugf("gettting posts for '%s'", p.User)
		envelopes, err = f.db.GetBasicPostsForUser(p.User)
	} else if p.Search != "" {

	} else {
		f.logger.Log.Debug("getting all envelopes")
		// return all envelopes
		envelopes, err = f.db.GetBasicPosts()
	}
	if err != nil {
		return
	}
	posts = make([]Post, len(envelopes))
	i := 0
	for _, e := range envelopes {
		posts[i] = f.MakePostWithComments(e)
		i++
	}
	posts = posts[:i]
	f.logger.Log.Info(time.Since(t))
	return
}

func (f *Feed) MakePostWithComments(e letter.Envelope) (post Post) {
	// postInterface, found := f.caching.Get(e.ID)
	// if found {
	// 	f.logger.Log.Debug("using cache")
	// 	post = postInterface.(Post)
	// 	return
	// }
	basicPost := f.MakePost(e)
	post = Post{
		Post:     basicPost,
		Comments: f.DetermineComments(basicPost.ID),
	}
	// f.caching.Set(e.ID, post, 3*time.Second)
	return
}

// This is needed for http rest api
func (self Feed) GetDatabase() database.DatabaseAPI {
	return self.db
}

//
// func (self Feed) ShowPostsForApi() ([]database.ApiBasicPost, error) {
// 	posts, err := self.db.GetPostsForApi()
// 	return posts, err
// }
//
// func (self Feed) ShowPostCommentsForApi(post_id string) ([]database.ApiBasicPost, error) {
// 	posts, err := self.db.GetPostCommentsForApi(post_id)
// 	return posts, err
// }
//
// func (self Feed) ShowPostForApi(post_id string) ([]database.ApiBasicPost, error) {
// 	posts, err := self.db.GetPostForApi(post_id)
// 	return posts, err
// }
//
// func (self Feed) ShowUserForApi(user_id string) (database.ApiUser, error) {
// 	user, err := self.db.GetUserForApi(user_id)
// 	return user, err
// }

func (f *Feed) MakePost(e letter.Envelope) (post BasicPost) {
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
		Likes: f.db.NumberOfLikes(e.ID),
	}
	return
}

func (f *Feed) DetermineComments(postID string) []BasicPost {
	return f.recurseComments(postID, []BasicPost{}, 0)
}

func (f *Feed) recurseComments(postID string, comments []BasicPost, depth int) []BasicPost {
	es, err := f.db.GetReplies(postID)
	if err != nil {
		f.logger.Log.Error(err)
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
func (f *Feed) AddFriendsKey() (err error) {
	// generate a key for friends
	myfriends := keypair.New()
	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey, making keypair")
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
		err = errors.Wrap(err, "AddFriendsKey, processing letter")
		return
	}

	if err != nil {
		err = errors.Wrap(err, "AddFriendsKey, processing public letter")
		return
	}

	return
}

// GetEnvelope will return an envelope with the given ID
func (f *Feed) GetEnvelope(id string) (e letter.Envelope, err error) {
	return f.db.GetEnvelopeFromID(id)
}

// GetIDs will return an envelope with the given ID
func (f *Feed) GetIDs() (ids map[string]struct{}, err error) {
	return f.db.GetIDs()
}

// GetConnected returns the users that are currently connected to
func (f *Feed) GetConnected() (us []User) {
	f.servers.RLock()
	defer f.servers.RUnlock()
	i := 0
	us = make([]User, len(f.servers.connected))
	for address := range f.servers.connected {
		us[i] = f.servers.connected[address]
	}
	return
}

// Sync will try to sync with the respective address
func (f *Feed) Sync(address string) (err error) {
	f.logger.Log.Debugf("syncing with %s", address)

	// get the information about the kiki server
	err = f.PingKikiInstance(address)
	if err != nil {
		return errors.Wrap(err, "syncing ping doesn't work")
	}

	// Get a list of my IDs
	myIDs, err := f.GetIDs()
	if err != nil {
		return
	}

	// get the list
	var target Response
	req, err := http.NewRequest("GET", address+"/list", nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return
	}
	if "ok" != target.Status {
		return errors.New(target.Error)
	}

	f.logger.Log.Debugf("got %d IDs from %s", len(target.IDs), address)

	// check whether I need any of their envelopes
	for theirID := range target.IDs {
		if _, ok := myIDs[theirID]; ok {
			continue
		}
		f.logger.Log.Debugf("%s has new envelope: %s", address, theirID)
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
		f.logger.Log.Debugf("my envelope %s is new to %s", myID, address)
		err = f.UploadEnvelope(address, myID)
		if err != nil {
			return
		}
	}

	f.UpdateEverything()
	return
}

// UploadEnvelope will upload the specified envelope
func (f *Feed) UploadEnvelope(address, id string) (err error) {
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

	f.logger.Log.Debugf("uploaded %s to %s", id, address)

	return
}

// DownloadEnvelope will download the specified envelope
func (f *Feed) DownloadEnvelope(address, id string) (err error) {
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

	f.logger.Log.Debugf("downloaded %s from %s", target.Envelope.ID, address)

	err = f.ProcessEnvelope(target.Envelope)
	if err != nil {
		f.logger.Log.Error(err)
	}
	return
}

// PingKikiInstance will ping a kiki instance to see if it is viable
func (f *Feed) PingKikiInstance(address string) (err error) {
	f.servers.RLock()
	if _, ok := f.servers.connected[address]; ok {
		f.logger.Log.Debugf("already connected to %s", address)
		f.servers.RUnlock()
		return
	}
	f.servers.RUnlock()

	timeout := time.Duration(1500 * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
	}

	regionSignature, _ := f.RegionKey.Signature(f.RegionKey)
	personalSignature, _ := f.PersonalKey.Signature(f.RegionKey)
	payload := Response{
		PersonalPublicKey: f.PersonalKey.Public,
		PersonalSignature: personalSignature,
		RegionPublicKey:   f.RegionKey.Public,
		RegionSignature:   regionSignature,
	}
	bPayload, _ := json.Marshal(payload)
	body := bytes.NewReader(bPayload)
	f.logger.Log.Debugf("POST %s/handshake", address)
	resp, err := client.Post(address+"/handshake", "application/json", body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	f.logger.Log.Debug("decoding response")
	var target Response
	err = json.NewDecoder(resp.Body).Decode(&target)
	if err != nil {
		return
	}
	if "ok" != target.Status {
		f.logger.Log.Debug("got error")
		err = errors.New(target.Error)
		return
	}

	f.logger.Log.Debugf("validating %s", address)
	err = f.ValidateKikiInstance(target)
	if err == nil {
		f.AddAddressToServers(address, target)
	} else {
		f.logger.Log.Warnf("could not validate %s", address)
	}
	return
}

// ValidateKikiInstance will validate whether a ping response is valid when POSTing or when recieving
func (f *Feed) ValidateKikiInstance(r Response) (err error) {
	// validate that the same region sent the signature
	err = f.RegionKey.Validate(r.RegionSignature, f.RegionKey)
	if err != nil {
		f.logger.Log.Warn(err)
		err = errors.Wrap(err, "could not validate region key")
		return
	}
	senderKey, err := keypair.FromPublic(r.PersonalPublicKey)
	if err != nil {
		f.logger.Log.Warn(err)
		err = errors.Wrap(err, "problem deciphering key")
		return
	}
	err = f.RegionKey.Validate(r.PersonalSignature, senderKey)
	if err != nil {
		f.logger.Log.Warn(err)
		err = errors.Wrap(err, "could not validate personal key")
		return
	}
	return
}

func (f *Feed) AddAddressToServers(address string, r Response) {
	u := f.GetUser(r.PersonalPublicKey)
	u.Server = address

	f.servers.Lock()
	defer f.servers.Unlock()
	f.logger.Log.Infof("connected to new server %s: %+v", address, u)
	f.servers.connected[address] = u
	alreadyRecorded := false
	for _, currentAddress := range f.Settings.AvailableServers {
		if address == currentAddress {
			alreadyRecorded = true
			break
		}
	}
	if !alreadyRecorded {
		f.Settings.AvailableServers = append(f.Settings.AvailableServers, address)
		f.Save()
	}
}

// PurgeOverflowingStorage will delete old messages
func (f *Feed) PurgeOverflowingStorage() (err error) {
	users, err := f.db.ListUsers()
	if err != nil {
		return
	}
	_, _, friendsList := f.db.Friends(f.PersonalKey.Public)
	friendsMap := make(map[string]struct{})
	for _, friend := range friendsList {
		friendsMap[friend] = struct{}{}
	}

	for _, user := range users {
		// skip personal user
		if user == f.PersonalKey.Public {
			continue
		}

		// determine limit
		limit := f.Settings.StoragePerPublicPerson
		if _, ok := friendsMap[user]; ok {
			limit = f.Settings.StoragePerFriend
		}

		currentSpace, err2 := f.db.DiskSpaceForUser(user)
		if err2 != nil {
			return err2
		}

		// don't proceed if the current space does not exceed
		if currentSpace < limit {
			continue
		}

		// first purge repeated actions (changing names multiple times)
		err2 = f.db.DeleteOldActions(user)
		if err2 != nil {
			return err2
		}

		// then purge edits
		if currentSpace > limit {
			err2 = f.db.DeleteUsersEdits(user)
			if err2 != nil {
				return err2
			}
		}

		for {
			f.logger.Log.Infof("user: %s: space: %d / %d", user, currentSpace, limit)
			if currentSpace < limit {
				break
			}
			err = f.db.DeleteUsersOldestPost(user)
			if err != nil {
				return
			}
			currentSpace, err2 = f.db.DiskSpaceForUser(user)
			if err2 != nil {
				return err2
			}
		}
	}
	return
}

func (f *Feed) TestStuff() {
	err := f.PingKikiInstance("http://localhost:8005")
	f.logger.Log.Error(err.Error())
	f.logger.Log.Error(err == nil)
}
