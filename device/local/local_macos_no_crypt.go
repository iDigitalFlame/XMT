//go:build (darwin || ios) && !crypt

package local

import (
	"os"
	"os/exec"
	"os/user"
	"strings"

	"golang.org/x/sys/unix"
)

func sysID() []byte {
	o, err := exec.Command("/usr/sbin/ioreg", "-rd1", "-c", "IOPlatformExpertDevice").CombinedOutput()
	if err != nil || len(o) == 0 {
		return nil
	}
	for _, v := range strings.Split(string(o), "\n") {
		if !strings.Contains(v, `"IOPlatformUUID"`) {
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
func uname() string {
	var (
		u   unix.Utsname
		err = unix.Uname(&u)
	)
	if err != nil {
		return ""
	}
	var (
		v = make([]byte, 65)
		i int
	)
	for ; i < 65; i++ {
		if u.Release[i] == 0 {
			break
		}
		v[i] = byte(u.Release[i])
	}
	return string(v[:i])
}
func version() string {
	var b, n, v string
	if o, err := exec.Command("/usr/bin/sw_vers").CombinedOutput(); err == nil {
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
		n = m["PRODUCTNAME"]
		b = m["BUILDVERSION"]
		v = m["PRODUCTVERSION"]
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "MacOS"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "MacOS (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "MacOS (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "MacOS (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "MacOS"
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
