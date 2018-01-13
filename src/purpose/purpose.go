package purpose

// There are the available purposes for a letter
var (
	// Share text will share a text post
	ShareText = "share-text"

	// SharePNG shares a PNG image
	SharePNG = "share-image/png"

	// ShareJPG shares a JPG image
	ShareJPG = "share-image/jpg"

	// ShareKey is for sharing keypairs with friends, or with self
	ShareKey = "share-key"

	// Actionments are always public

	// ActionFollow will follow someone
	ActionFollow = "action-follow"

	// ActionLike will give a person a like
	ActionLike = "action-like"

	// ActionName will assign a person a name
	ActionName = "action-assign/name"

	ActionProfile = "action-assign/profile"

	ActionImage = "action-assign/image"

	ActionBlock = "action-assign/block"

	// ActionErase will erase a persons profile from every carrier
	ActionErase = "action-erase"
)

func Valid(purpose string) bool {
	for _, p := range []string{ShareJPG, SharePNG, ShareText, ShareKey, ActionFollow, ActionName, ActionBlock, ActionProfile, ActionLike, ActionImage} {
		if purpose == p {
			return true
		}
	}
	return false
}
