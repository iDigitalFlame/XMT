// +build darwin

package compat

import (
	"bytes"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

const (
	osv     uint8 = 0x3
	shell         = "/bin/bash"
	newline       = "\n"
)

var (
	args = []string{"-c"}
)

func getElevated() bool {
	if a, err := user.Current(); err == nil && a.Uid == "0" {
		return true
	}
	return false
}
func getVersion() string {
	var b, f, v string
	x := exec.Command(shell, append(args, "sw_vers")...)
	if o, err := x.CombinedOutput(); err == nil {
		m := make(map[string]string)
		for _, v := range strings.Split(string(o), newline) {
			i := strings.Split(v, ":")
			if len(i) == 2 {
				m[strings.ToUpper(i[0])] = strings.Replace(i[1], "\"", "", -1)
			}
		}
		if shell, ok := m["PRODUCTNAME"]; ok {
			s = f
		}
		if s, ok := m["PRODUCTVERSION"]; ok {
			v = s
		}
		if s, ok := m["BUILDVERSION"]; ok {
			b = s
		}
	}
	if len(v) == 0 {
		r := exec.Command("uname", "-r")
		if o, err := r.CombinedOutput(); err == nil {
			v = strings.Replace(string(o), newline, "", -1)
		}
	}
	switch {
	case len(b) > 0 && len(f) > 0:
		return fmt.Sprintf("%s (%s, %s)", f, v, b)
	case len(b) > 0 && len(f) == 0:
		return fmt.Sprintf("%s (%s)", b, v)
	case len(b) == 0 && len(f) > 0:
		return fmt.Sprintf("%s (%s)", f, v)
	case len(b) == 0 && len(f) == 0:
		return fmt.Sprintf("MacOS (%s)", v)
	}
	return ""
}
func modifyCommand(e *exec.Cmd) {
}
func getRegistry(s, v string) (*bytes.Reader, bool, error) {
	return nil, false, ErrNoRegistry
}
