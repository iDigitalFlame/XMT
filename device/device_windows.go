// +build windows

package device

import (
	"os"
	"strconv"

	"golang.org/x/sys/windows/registry"
)

const (
	// OS is the local machine's Operating System type.
	OS = Windows

	// Newline is the machine specific newline character.
	Newline = "\n"
)

var (
	// Shell is the default machine specific command shell.
	Shell = shell()

	// ShellArgs is the default machine specific command shell arguments to run commands.
	ShellArgs = []string{"/c"}
)

func shell() string {
	if s, ok := os.LookupEnv("ComSpec"); ok {
		return s
	}
	if d, ok := os.LookupEnv("WinDir"); ok {
		p := d + `\system32\cmd.exe`
		if s, err := os.Stat(p); err == nil && !s.IsDir() {
			return p
		}
	}
	return `%WinDir%\system32\cmd.exe`
}
func isElevated() bool {
	if p, err := os.OpenFile(`\\.\PHYSICALDRIVE0`, os.O_RDONLY, 0); err == nil {
		p.Close()
		return true
	}
	return false
}
func getVersion() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "Windows (?)"
	}
	var (
		b, v    string
		n, _, _ = k.GetStringValue("ProductName")
	)
	if s, _, err := k.GetStringValue("CurrentBuild"); err == nil {
		b = s
	} else if s, _, err := k.GetStringValue("ReleaseId"); err == nil {
		b = s
	}
	if i, _, err := k.GetIntegerValue("CurrentMajorVersionNumber"); err == nil {
		if x, _, err := k.GetIntegerValue("CurrentMinorVersionNumber"); err == nil {
			v = strconv.Itoa(int(i)) + "." + strconv.Itoa(int(x))
		} else {
			v = strconv.Itoa(int(i))
		}
	} else {
		v, _, _ = k.GetStringValue("CurrentVersion")
	}
	switch k.Close(); {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Windows (?)"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return "Windows (" + v + ", " + b + ")"
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return "Windows (" + v + ")"
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return "Windows (" + b + ")"
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return n + " (" + v + ", " + b + ")"
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return n + " (" + v + ")"
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return n + " (" + b + ")"
	}
	return "Windows (?)"
}
