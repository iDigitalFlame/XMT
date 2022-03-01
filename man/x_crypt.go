//go:build crypt

package man

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	local = crypt.Get(11) // localhost:
	execA = crypt.Get(12) // *.so
	execB = crypt.Get(13) // *.dll
	execC = crypt.Get(14) // *.exe
)

func (o objSync) String() string {
	switch o {
	case Mutex:
		return crypt.Get(7) // mutex
	case Event:
		return crypt.Get(8) // event
	case Mailslot:
		return crypt.Get(9) // mailslot
	case Semaphore:
		return crypt.Get(10) // semaphore
	}
	return crypt.Get(7) // mutex
}
