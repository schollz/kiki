package utils

import (
	"strings"
	"time"

	"github.com/ararog/timeago"
)

func TimeAgo(t time.Time) string {
	got, _ := timeago.TimeAgoWithTime(time.Now(), t)
	return strings.ToLower(got)
}
