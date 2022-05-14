//go:build plan9 && crypt

package local

import (
	"os"
	"os/exec"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	if b, err := os.ReadFile(crypt.Get(107)); err == nil { // /etc/hostid
		return b
	}
	// kenv -q smbios.system.uuid
	o, _ := exec.Command(crypt.Get(108), crypt.Get(109), crypt.Get(110)).CombinedOutput()
	return o
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
		v = crypt.Get(81) // plan9
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(81) // plan9
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(81) + " (" + v + ", " + b + ")" // plan9
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(81) + " (" + v + ")" // plan9
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(81) + " (" + b + ")" // plan9
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(81) // plan9
}
