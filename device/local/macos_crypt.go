//go:build (darwin || ios) && crypt

package local

import (
	"os/exec"
	"strings"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

func sysID() []byte {
	// /usr/sbin/ioreg -rd1 -c IOPlatformExpertDevice
	o, err := exec.Command(crypt.Get(82), crypt.Get(83), crypt.Get(35), crypt.Get(84)).CombinedOutput()
	if err != nil || len(o) == 0 {
		return nil
	}
	for _, v := range strings.Split(string(o), "\n") {
		if !strings.Contains(v, crypt.Get(85)) { // IOPlatformUUID
			continue
		}
		x := strings.IndexByte(v, '=')
		if x < 14 || len(v)-x < 2 {
			continue
		}
		c, s := len(v)-1, x+1
		for ; c > x && (v[c] == '"' || v[c] == ' ' || v[c] == 9); c-- {
		}
		for ; s < c && (v[s] == '"' || v[s] == ' ' || v[s] == 9); s++ {
		}
		if s == c || s > len(v) || s > c {
			continue
		}
		return []byte(v[s : c+1])
	}
	return nil
}
func version() string {
	var b, n, v string
	if o, err := exec.Command(crypt.Get(86)).CombinedOutput(); err == nil { // /usr/bin/sw_vers
		m := make(map[string]string)
		for _, v := range strings.Split(string(o), "\n") {
			x := strings.IndexByte(v, ':')
			if x < 1 || len(v)-x < 2 {
				continue
			}
			c, s := len(v)-1, x+1
			for ; c > x && (v[c] == ' ' || v[c] == 9); c-- {
			}
			for ; s < c && (v[s] == ' ' || v[s] == 9); s++ {
			}
			m[strings.ToUpper(v[:x])] = v[s : c+1]
		}
		n = m[crypt.Get(87)] // PRODUCTNAME
		b = m[crypt.Get(88)] // BUILDVERSION
		v = m[crypt.Get(89)] // PRODUCTVERSION
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return crypt.Get(23) // MacOS
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return crypt.Get(23) + " (" + v + ", " + b + ")" // MacOS
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return crypt.Get(23) + " (" + v + ")" // MacOS
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return crypt.Get(23) + " (" + b + ")" // MacOS
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return crypt.Get(23) // MacOS
}
