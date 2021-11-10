//go:build !client
// +build !client

package c2

import (
	"strconv"
	"time"
)

const maxEvents = 2048

// String returns the details of this Session as a string.
func (s *Session) String() string {
	switch {
	case s.parent == nil && s.sleep == 0:
		return "[" + s.ID.String() + "] -> " + s.host + " " + s.Last.Format(time.RFC1123)
	case s.parent == nil && (s.jitter == 0 || s.jitter > 100):
		return "[" + s.ID.String() + "] " + s.sleep.String() + " -> " + s.host
	case s.parent == nil:
		return "[" + s.ID.String() + "] " + s.sleep.String() + "/" + strconv.Itoa(int(s.jitter)) + "% -> " + s.host
	case s.parent != nil && (s.jitter == 0 || s.jitter > 100):
		return "[" + s.ID.String() + "] " + s.sleep.String() + " -> " + s.host + " " + s.Last.Format(time.RFC1123)
	}
	return "[" + s.ID.String() + "] " + s.sleep.String() + "/" + strconv.Itoa(int(s.jitter)) + "% -> " + s.host + " " + s.Last.Format(time.RFC1123)
}
