//go:build linux
// +build linux

package device

import (
	"os"
	"os/exec"
	"os/user"
	"strings"
)

const (
	// OS is the local machine's Operating System type.
	OS = Linux

	// Shell is the default machine specific command shell.
	Shell = "/bin/bash"
	// Newline is the machine specific newline character.
	Newline = "\n"
)

// ShellArgs is the default machine specific command shell arguments to run commands.
var ShellArgs = []string{"-c"}

func isElevated() bool {
	if os.Geteuid() == 0 || os.Getuid() == 0 {
		return true
	}
	if a, err := user.Current(); err == nil && a.Uid == "0" {
		return true
	}
	return false
}
func getVersion() string {
	var (
		ok      bool
		b, n, v string
	)
	if o, err := exec.Command(Shell, append(ShellArgs, "cat /etc/*release*")...).CombinedOutput(); err == nil {
		m := make(map[string]string)
		for _, v := range strings.Split(string(o), Newline) {
			if i := strings.Split(v, "="); len(i) == 2 {
				m[strings.ToUpper(i[0])] = strings.ReplaceAll(i[1], `"`, "")
			}
		}
		b = m["ID"]
		if n, ok = m["PRETTY_NAME"]; !ok {
			n = m["NAME"]
		}
		if v, ok = m["VERSION_ID"]; !ok {
			v = m["VERSION"]
		}
	}
	if len(v) == 0 {
		if o, err := exec.Command("uname", "-r").CombinedOutput(); err == nil {
			v = strings.ReplaceAll(string(o), Newline, "")
		}
	}
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Linux (?)"
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
	return "Linux (?)"
}
