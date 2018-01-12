package feed

import (
	"html/template"
	"time"

	"github.com/cihub/seelog"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/keypair"
)

// Feed stores your basic data
type Feed struct {
	RegionKey   keypair.KeyPair `json:"region_key"`
	Settings    Settings        `json:"settings"`
	PersonalKey keypair.KeyPair `json:"personal_key"`

	storagePath string
	db          database.DatabaseAPI
	log         seelog.LoggerInterface
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
	Post     BasicPost
	Comments []BasicPost
}

type BasicPost struct {
	Depth      int
	ID         string
	Recipients string
	ReplyTo    string
	Content    template.HTML
	Date       time.Time
	TimeAgo    string
	User       User
}

type User struct {
	Name      string
	Profile   template.HTML
	PublicKey string
	Image     string
}
