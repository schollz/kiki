package purpose

// There are the available purposes for a letter
var (
	ShareText = "share-text"
	SharePNG  = "share-image/png"
	ShareJPG  = "share-image/jpg"
	// ShareKey is for sharing keypairs with friends, or with self
	ShareKey     = "share-key"
	AssignFollow = "assign-follow"
	AssignName   = "assign-name"
	AssignBlock  = "assign-block"
)

func Valid(purpose string) bool {
	for _, p := range []string{ShareJPG, SharePNG, ShareText, ShareKey, AssignFollow, AssignName, AssignBlock} {
		if purpose == p {
			return true
		}
	}
	return false
}
