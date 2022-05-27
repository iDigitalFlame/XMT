//go:build crypt

package man

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	local = crypt.Get(108) // localhost:
	execA = crypt.Get(1)   // *.so
	execB = crypt.Get(2)   // *.dll
	execC = crypt.Get(3)   // *.exe
)

func (o objSync) String() string {
	switch o {
	case Mutex:
		return crypt.Get(109) // mutex
	case Event:
		return crypt.Get(110) // event
	case Mailslot:
		return crypt.Get(111) // mailslot
	case Semaphore:
		return crypt.Get(112) // semaphore
	}
	return crypt.Get(109) // mutex
}
