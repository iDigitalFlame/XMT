//go:build (android || linux) && !crypt

package local

import (
	"os"
	"os/user"
)

func sysID() []byte {
	b, err := os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil {
		return b
	}
	b, _ = os.ReadFile("/etc/machine-id")
	return b
}
func version() string {
	var (
		ok      bool
		b, n, v string
	)
	if m := release(); len(m) > 0 {
		b = m["ID"]
		if n, ok = m["PRETTY_NAME"]; !ok {
			n = m["NAME"]
		}
		if v, ok = m["VERSION_ID"]; !ok {
			v = m["VERSION"]
		}
	}
	if len(v) == 0 {
		v = uname()
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Linux"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "Linux (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "Linux (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "Linux (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "Linux"
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
