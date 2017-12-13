package block

type Blob struct {
	ID         string   // hash of sender + data
	OriginalID string   // original ID, different than ID if overwriting
	Data       string   // base64 encoded bytes of data
	Action     string   // action verb
	Kind       string   // kind of action
	Channels   []string // channels for showing the post
	ReplyTo    string   // hash that Blob is response to
}
