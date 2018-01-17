package database

import (
	"encoding/json"
)

type ApiBasicPost struct {
	ID          string   `json:"id"`
	Recipients  []string `json:"recipients"`
	ReplyTo     string   `json:"reply_to"`
	Content     string   `json:"content"`
	Timestamp   int64    `json:"timestamp"`
	OwnerId     string   `json:"owner_id"`
	OwnerName   string   `json:"owner_name"`
	Likes       int64    `json:"likes"`
	NumComments int64    `json:"num_comments"`
	Purpose     string   `json:"purpose,omitempty"`
}

func (self *ApiBasicPost) Unmarshal(text string) error {
	err := json.Unmarshal([]byte(text), &self)
	return err
}

type ApiUser struct {
	Name      string   `json:"name"`
	PublicKey string   `json:"public_key"`
	Profile   string   `json:"profile"`
	Image     string   `json:"image"`
	Followers []string `json:"followers"`
	Following []string `json:"following"`
	Blocked   []string `json:"blocked"`
	// Friends
}

func (self *ApiUser) Unmarshal(text string) error {
	err := json.Unmarshal([]byte(text), &self)
	return err
}
