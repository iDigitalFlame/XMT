//go:build plan9 && !crypt

package local

import (
	"os"
	"os/exec"
)

func sysID() []byte {
	if b, err := os.ReadFile("/etc/hostid"); err == nil {
		return b
	}
	o, _ := exec.Command("kenv", "-q", "smbios.system.uuid").CombinedOutput()
	return o
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
		v = "plan9"
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "plan9"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "plan9 (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "plan9 (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "plan9 (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "plan9"
}
