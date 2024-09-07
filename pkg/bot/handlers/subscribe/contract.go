package subscribe

import (
	"regexp"
)

type publishedMessage struct {
	Id      string
	Time    int64
	Expires int64
	Event   string
	Topic   string
	Message string
}

// from ntfy.sh source code - https://github.com/binwiederhier/ntfy/blob/72f36f8296aec9c67a14dbff459e801d63084635/user/types.go#L245
var allowedTopicRegex = regexp.MustCompile(`^[-_A-Za-z0-9]{1,64}$`) // No '*'
