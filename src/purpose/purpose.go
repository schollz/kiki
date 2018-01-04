package purpose

// There are the available purposes for a letter
var (
	ShareText    = "share-text"
	SharePNG     = "share-image/png"
	ShareJPG     = "share-image/jpg"
	AssignFollow = "assign-follow"
)

func Valid(purpose string) bool {
	for _, p := range []string{ShareJPG, SharePNG, ShareText, AssignFollow} {
		if purpose == p {
			return true
		}
	}
	return false
}
