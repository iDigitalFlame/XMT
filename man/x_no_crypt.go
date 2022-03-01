//go:build !crypt

package man

const (
	local = "localhost:"
	execA = "*.so"
	execB = "*.dll"
	execC = "*.exe"
)

func (o objSync) String() string {
	switch o {
	case Mutex:
		return "mutex"
	case Event:
		return "event"
	case Mailslot:
		return "mailslot"
	case Semaphore:
		return "semaphore"
	}
	return "mutex"
}
