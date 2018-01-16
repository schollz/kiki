package feed

import (
	"html/template"
	"sync"
	"time"

	"github.com/cihub/seelog"
	cache "github.com/robfig/go-cache"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
)

type Response struct {
	RegionSignature   string              `json:"region_signature"`
	RegionPublicKey   string              `json:"region_key"`
	PersonalSignature string              `json:"personal_signature"`
	PersonalPublicKey string              `json:"personal_key"`
	IDs               map[string]struct{} `json:"ids"`
	Envelope          letter.Envelope     `json:"envelope"`
	Error             string              `json:"error"`
	Message           string              `json:"message"`
	Status            string              `json:"status"`
}

// Feed stores your basic data
type Feed struct {
	RegionKey   keypair.KeyPair `json:"region_key"`
	Settings    Settings        `json:"settings"`
	PersonalKey keypair.KeyPair `json:"personal_key"`

	storagePath string
	db          database.DatabaseAPI
	log         seelog.LoggerInterface
	logger      logging.SeelogWrapper
	caching     *cache.Cache
	servers     connections
}

type connections struct {
	connected map[string]User
	sync.RWMutex
}

type Settings struct {
	StoragePerPublicPerson int64    `json:"storage_per_person"`  // maximum size in bytes to store of public messages. Once exceeded, old messages are purged
	StoragePerFriend       int64    `json:"storage_per_friend"`  // maximum size in bytes to store of friend messages. Once exceeded, old messages are purged
	FriendsOfFriends       bool     `json:"friends_of_friends"`  // whether you want to share your friends friend keys with new friends, effectively making a new friend friends with all your friends. This also means that when you make a new friend, that friends key is emitted to all your current friends. (default: true)
	BlockPublicPhotos      bool     `json:"block_public_photos"` // if true, block the transfer of any public photos to your computer
	HidePublicPosts        bool     `json:"hide_public_posts"`   // if true, works as a diary basically
	AvailableServers       []string `json:"available_servers"`
}

// GenerateSettings create new instance of Something
func GenerateSettings() Settings {
	return Settings{
		StoragePerPublicPerson: 5000000,  // 5 MB
		StoragePerFriend:       50000000, // 50 MB
		FriendsOfFriends:       true,
		BlockPublicPhotos:      false,
		HidePublicPosts:        false,
		AvailableServers:       []string{"https://kiki.network"},
	}
}

type Post struct {
	Post     BasicPost   `json:"post"`
	Comments []BasicPost `json:"comments"`
}

type BasicPost struct {
	Depth      int           `json:"depth"`
	ID         string        `json:"id"`
	Recipients string        `json:"recipients"`
	ReplyTo    string        `json:"reply_to"`
	Content    template.HTML `json:"content"`
	Date       time.Time     `json:"date"`
	TimeAgo    string        `json:"time_ago"`
	User       User          `json:"user"`
	Likes      int64         `json:"likes"`
	Comments   []BasicPost   `json:"comments"`
}

// // TESTING
// type ApiBasicPost struct {
// 	// Depth      int           `json:"depth"`
// 	ID          string        `json:"id"`
// 	Recipients  string        `json:"recipients"`
// 	ReplyTo     string        `json:"reply_to"`
// 	Content     template.HTML `json:"content"`
// 	Date        time.Time     `json:"date"`
// 	TimeAgo     string        `json:"time_ago"`
// 	OwnerId     string        `json:"owner_id"`
// 	Likes       int64         `json:"likes"`
// 	NumComments int64         `json:"num_comments"`
// }

type User struct {
	Name      string        `json:"name"`
	Profile   template.HTML `json:"profile"`
	PublicKey string        `json:"public_key"`
	Image     string        `json:"image"`
	Followers []string      `json:"followers"`
	Following []string      `json:"following"`
	Friends   []string      `json:"friends"`
}

type UserFriends struct {
	Friends   []User `json:"user_friends"`
	Followers []User `json:"user_followers"`
	Following []User `json:"user_following"`
}
