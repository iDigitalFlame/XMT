// +build windows

package device

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
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

	dllAdvapi32               = windows.NewLazySystemDLL("advapi32.dll")
	funcAdjustTokenPrivileges = dllAdvapi32.NewProc("AdjustTokenPrivileges")
)

type privileges struct {
	PrivilegeCount uint32
	Privileges     [5]windows.LUIDAndAttributes
}

func shell() string {
	if s, ok := os.LookupEnv("ComSpec"); ok {
		return s
	}
	if d, ok := os.LookupEnv("WinDir"); ok {
		p := fmt.Sprintf("%s\\system32\\cmd.exe", d)
		if s, err := os.Stat(p); err == nil && !s.IsDir() {
			return p
		}
	}
	return "%WinDir%\\system32\\cmd.exe"
}
func isElevated() bool {
	if p, err := os.Open("\\\\.\\PHYSICALDRIVE0"); err == nil {
		p.Close()
		return true
	}
	return false
}
func getVersion() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return Windows.String()
	}
	var b, v string
	n, _, _ := k.GetStringValue("ProductName")
	if s, _, err := k.GetStringValue("CurrentBuild"); err == nil {
		b = s
	} else if s, _, err := k.GetStringValue("ReleaseId"); err == nil {
		b = s
	}
	if i, _, err := k.GetIntegerValue("CurrentMajorVersionNumber"); err == nil {
		if x, _, err := k.GetIntegerValue("CurrentMinorVersionNumber"); err == nil {
			v = fmt.Sprintf("%d.%d", i, x)
		} else {
			v = strconv.Itoa(int(i))
		}
	} else {
		v, _, _ = k.GetStringValue("CurrentVersion")
	}
	k.Close()
	switch {
	case len(n) == 0 && len(b) == 0 && len(v) == 0:
		return "Windows (?)"
	case len(n) == 0 && len(b) > 0 && len(v) > 0:
		return fmt.Sprintf("Windows (%s, %s)", v, b)
	case len(n) == 0 && len(b) == 0 && len(v) > 0:
		return fmt.Sprintf("Windows (%s)", v)
	case len(n) == 0 && len(b) > 0 && len(v) == 0:
		return fmt.Sprintf("Windows (%s)", b)
	case len(n) > 0 && len(b) > 0 && len(v) > 0:
		return fmt.Sprintf("%s (%s, %s)", n, v, b)
	case len(n) > 0 && len(b) == 0 && len(v) > 0:
		return fmt.Sprintf("%s (%s)", n, v)
	case len(n) > 0 && len(b) > 0 && len(v) == 0:
		return fmt.Sprintf("%s (%s)", n, b)
	}
	return "Windows (?)"
}

// AdjustPrivileges will attempt to enable the supplied Windows privilege values on the current process's Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustPrivileges(s ...string) error {
	var t windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_WRITE|windows.TOKEN_QUERY, &t); err != nil {
		return fmt.Errorf("cannot get current token: %w", err)
	}
	err := AdjustTokenPrivileges(uintptr(t), s...)
	t.Close()
	return err
}
func adjust(h uintptr, s []string) error {
	var (
		n   *uint16
		p   privileges
		err error
	)
	for i := range s {
		if n, err = windows.UTF16PtrFromString(s[i]); err != nil {
			return fmt.Errorf("cannot convert %q: %w", s[i], err)
		}
		if err := windows.LookupPrivilegeValue(nil, n, &p.Privileges[i].Luid); err != nil {
			return fmt.Errorf("cannot lookup privilege %q: %w", s[i], err)
		}
		p.Privileges[i].Attributes = windows.SE_PRIVILEGE_ENABLED
	}
	p.PrivilegeCount = uint32(len(s))
	_, _, err = funcAdjustTokenPrivileges.Call(
		uintptr(h), 0,
		uintptr(unsafe.Pointer(&p)),
		uintptr(unsafe.Sizeof(p)),
		0, 0,
	)
	if e, ok := err.(syscall.Errno); ok && e == 0 {
		return nil
	}
	return fmt.Errorf("cannot assign all privileges: %w", err)
}

// Registry attempts to open a registry value or key, value pair on Windows devices. Returns err if the system is
// not a Windows device or an error occurred during the open. Always returns 'ErrNoWindows' on non-windows devices.
func Registry(key, value string) (*RegistryFile, error) {
	var k registry.Key
	switch p := strings.ToUpper(key); {
	case strings.HasPrefix(p, "HKEY_USERS") || strings.HasPrefix(p, "HKU"):
		k = registry.USERS
	case strings.HasPrefix(p, "HKEY_CURRENT_USER") || strings.HasPrefix(p, "HKCU"):
		k = registry.CURRENT_USER
	case strings.HasPrefix(p, "HKEY_CLASSES_ROOT") || strings.HasPrefix(p, "HKCR"):
		k = registry.CLASSES_ROOT
	case strings.HasPrefix(p, "HKEY_LOCAL_MACHINE") || strings.HasPrefix(p, "HKLM"):
		k = registry.LOCAL_MACHINE
	case strings.HasPrefix(p, "HKEY_CURRENT_CONFIG") || strings.HasPrefix(p, "HKCC"):
		k = registry.CURRENT_CONFIG
	case strings.HasPrefix(p, "HKEY_PERFORMANCE_DATA") || strings.HasPrefix(p, "HKPD"):
		k = registry.PERFORMANCE_DATA
	default:
		return nil, fmt.Errorf("registry path %q does not contain a valid key root", key)
	}
	i := strings.IndexRune(key, '\\')
	if i <= 0 {
		return nil, fmt.Errorf("registry path %q does not contain a valid key root", key)
	}
	h, err := registry.OpenKey(k, key[i+1:], registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	var y time.Time
	if m, err := h.Stat(); err == nil {
		y = m.ModTime()
	}
	if len(value) == 0 {
		return &RegistryFile{k: key, m: y}, h.Close()
	}
	defer h.Close()
	r, t, err := h.GetValue(value, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to read registry path \"%s:%s\": %w", key, value, err)
	}
	if r <= 0 {
		return nil, fmt.Errorf("registry path \"%s:%s\" returned a zero size", key, value)
	}
	b := make([]byte, r)
	if _, _, err := h.GetValue(value, b); err != nil {
		return nil, fmt.Errorf("unable to read registry path \"%s:%s\": %w", key, value, err)
	}
	var o io.Reader
	if t == registry.SZ || t == registry.EXPAND_SZ || t == registry.MULTI_SZ {
		o = strings.NewReader(windows.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(&b[0]))[: len(b)/2 : len(b)/2]))
	} else {
		o = bytes.NewReader(b)
	}
	return &RegistryFile{k: key, v: value, m: y, r: o}, nil
}

// AdjustTokenPrivileges will attempt to enable the supplied Windows privilege values on the supplied process Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustTokenPrivileges(h uintptr, s ...string) error {
	for x, w := 0, 0; x < len(s); {
		w = 5
		if x+w > len(s) {
			w = len(s) - x
		}
		if err := adjust(h, s[x:x+w]); err != nil {
			return err
		}
		x += w
	}
	return nil
}
