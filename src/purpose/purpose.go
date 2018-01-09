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
	// AssignFollow will follow someone
	AssignFollow = "assign-follow"
	// AssignLike will give a person a like
	AssignLike = "assign-like"
	// AssignName will assign a person a name
	AssignName    = "assign-name"
	AssignBlock   = "assign-block"
	AssignProfile = "assign-profile"
)

func Valid(purpose string) bool {
	for _, p := range []string{ShareJPG, SharePNG, ShareText, ShareKey, AssignFollow, AssignName, AssignBlock, AssignProfile, AssignLike} {
		if purpose == p {
			return true
		}
	}
	return false
}
