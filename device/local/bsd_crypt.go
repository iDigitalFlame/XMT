//go:build (aix || dragonfly || freebsd || hurd || illumos || nacl || netbsd || openbsd || plan9 || solaris || zos) && crypt

package local

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	switch runtime.GOOS {
	case "aix":
		// AIX specific support: https://github.com/denisbrodbeck/machineid/pull/16
		// lsattr -l  sys0 -a os_uuid -E
		if b, err := exec.Command(crypt.Get(98), crypt.Get(99), crypt.Get(100), crypt.Get(101), crypt.Get(102), crypt.Get(103)).CombinedOutput(); err == nil {
			if i := bytes.IndexByte(b, ' '); i > 0 {
				return b[i+1:]
			}
			return b
		}
	case "openbsd":
		// Support get hardware UUID for OpenBSD: https://github.com/denisbrodbeck/machineid/pull/14
		// sysctl -n hw.uuid
		if b, err := exec.Command(crypt.Get(104), crypt.Get(105), crypt.Get(106)).CombinedOutput(); err == nil {
			return b
		}
	}
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
	if len(b) == 0 && strings.Contains(runtime.GOOS, crypt.Get(111)) { // bsd
		if o, err := exec.Command(crypt.Get(112), crypt.Get(113)).CombinedOutput(); err == nil { // freebsd-version -k
			b = strings.ReplaceAll(string(o), "\n", "")
		}
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(114) // BSD
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(114) + " (" + v + ", " + b + ")" // BSD
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(114) + " (" + v + ")" // BSD
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(114) + " (" + b + ")" // BSD
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(114) // BSD
}
