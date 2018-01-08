package feed

import (
	"github.com/cihub/seelog"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/keypair"
)

type Feed struct {
	StoragePath string          `json:"storage_page"`
	RegionKey   keypair.KeyPair `json:"region_key"`
	Settings    Settings        `json:"settings"`
	personalKey keypair.KeyPair
	db          database.DatabaseAPI
	log         seelog.LoggerInterface
}

type Settings struct {
	StoragePerPublicPerson int64 // maximum size in bytes to store of public messages. Once exceeded, old messages are purged
	FriendsOfFriends       bool  // whether you want to share your friends friend keys with new friends, effectively making a new friend friends with all your friends. This also means that when you make a new friend, that friends key is emitted to all your current friends. (default: true)
	ShowPublicPhotos       bool  // if true, automatically show the display public photos (default: false)
	HidePublicPosts        bool  // if true, works as a diary basically
}

// GenerateSettings create new instance of Something
func GenerateSettings() Settings {
	return Settings{
		StoragePerPublicPerson: 5000000, // 5 MB
		FriendsOfFriends:       true,
		ShowPublicPhotos:       true,
	}
}
