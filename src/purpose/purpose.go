package purpose

// There are the available purposes for a letter
var (
	// Share text will share a text post
	// Content: HTML
	ShareText = "share-text"

	// SharePNG shares a PNG image
	// Content: Base64 of image
	SharePNG = "share-image/png"

	// ShareJPG shares a JPG image
	// Content: Base64 of image
	ShareJPG = "share-image/jpg"

	// ShareKey is for sharing keypairs with friends, or with self
	// Content: Marshalled keypair.KeyPair
	ShareKey = "share-key"

	// Actions are always public

	// ActionFollow will follow someone
	// Content: Public key of the person being followed
	ActionFollow = "action-follow"

	// ActionLike will give a person a like
	// Content: ID of the post being liked
	ActionLike = "action-like"

	// ActionName will assign a person a name
	// Content: Text of name
	ActionName = "action-assign/name"

	// ActionProfile will assign a person a profile
	// Content: HTML of profile
	ActionProfile = "action-assign/profile"

	// ActionProfile will assign a person a profile image
	// Content: ID of the image letter
	ActionImage = "action-assign/image"

	// ActionBlock will block a person
	// Content: Public key of the person to block
	ActionBlock = "action-block"

	// ActionErase will erase a persons profile from every carrier
	// Content: Empty
	ActionErase = "action-erase"
)

func Valid(purpose string) bool {
	for _, p := range []string{ShareJPG, SharePNG, ShareText, ShareKey, ActionFollow, ActionName, ActionBlock, ActionProfile, ActionLike, ActionImage, ActionErase} {
		if purpose == p {
			return true
		}
	}
	return false
}
