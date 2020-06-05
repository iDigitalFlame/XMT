// +build linux

package device

import (
	"fmt"
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

var (
	// ShellArgs is the default machine specific command shell arguments to run commands.
	ShellArgs = []string{"-c"}
)

func isElevated() bool {
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
		return fmt.Sprintf("Linux (%s, %s)", v, b)
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return fmt.Sprintf("Linux (%s)", v)
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return fmt.Sprintf("Linux (%s)", b)
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return fmt.Sprintf("%s (%s, %s)", n, v, b)
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return fmt.Sprintf("%s (%s)", n, v)
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return fmt.Sprintf("%s (%s)", n, b)
	}
	return "Linux (?)"
}

// AdjustPrivileges will attempt to enable the supplied Windows privilege values on the current process's Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustPrivileges(_ ...string) error {
	return ErrNoWindows
}

// Registry attempts to open a registry value or key, value pair on Windows devices. Returns err if the system is
// not a Windows device or an error occurred during the open. Always returns 'ErrNoWindows' on non-windows devices.
func Registry(_, _ string) (*RegistryFile, error) {
	return nil, ErrNoWindows
}

// AdjustTokenPrivileges will attempt to enable the supplied Windows privilege values on the supplied process Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustTokenPrivileges(_ uintptr, _ ...string) error {
	return ErrNoWindows
}
