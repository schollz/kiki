package block

import "time"

type Block struct {
	Name       string    // hash of timestamp and payload
	Sender     string    // public key of the sender
	Recipients []string  // secret encrypted by each recipient public key
	Payload    string    // encrypted compressed Blob
	Timestamp  time.Time // time
}
