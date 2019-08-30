// +build windows

package compat

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	osv     uint8 = 0x0
	newline       = "\r\n"
)

var (
	args  = []string{"/c"}
	shell = "%WinDir%\\system32\\cmd.exe"
)

func init() {
	if p, ok := os.LookupEnv("ComSpec"); ok {
		shell = p
	} else {
		if l, ok := os.LookupEnv("WinDir"); ok {
			p := fmt.Sprintf("%s\\system32\\cmd.exe", l)
			if s, err := os.Stat(p); err == nil && !s.IsDir() {
				shell = p
			}
		}
	}
}
func getElevated() bool {
	if p, err := os.Open("\\\\.\\PHYSICALDRIVE0"); err == nil {
		p.Close()
		return true
	}
	return false
}
func getVersion() string {
	var b, f, v string
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE); err == nil {
		defer k.Close()
		if s, _, err := k.GetStringValue("ProductName"); err == nil {
			f = s
		}
		if s, _, err := k.GetStringValue("CurrentBuild"); err == nil {
			b = s
		}
		if len(b) == 0 {
			if s, _, err := k.GetStringValue("ReleaseId"); err == nil {
				b = s
			}
		}
		i, _, err := k.GetIntegerValue("CurrentMajorVersionNumber")
		if err != nil {
			if s, _, err := k.GetStringValue("CurrentVersion"); err == nil {
				v = s
			}
		} else {
			if x, _, err := k.GetIntegerValue("CurrentMinorVersionNumber"); err == nil {
				v = fmt.Sprintf("%d.%d", i, x)
			} else {
				v = strconv.Itoa(int(i))
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
			return fmt.Sprintf("Windows (%s)", v)
		}
	}
	return ""
}
func modifyCommand(e *exec.Cmd) {
	if strings.HasSuffix(e.Args[0], "cmd.exe") {
		e.SysProcAttr = &syscall.SysProcAttr{
			CmdLine:       strings.Join(e.Args, " "),
			HideWindow:    false,
			CreationFlags: 0,
		}
	}
}
func getRegistry(s, v string) (*bytes.Reader, bool, error) {
	p := strings.ToUpper(s)
	k := registry.LOCAL_MACHINE
	if strings.HasPrefix(p, "HKEY_CURRENT_USER") || strings.HasPrefix(p, "HKCU") {
		k = registry.CURRENT_USER
	} else if strings.HasPrefix(p, "HKEY_CLASSES_ROOT") || strings.HasPrefix(p, "HKCR") {
		k = registry.CLASSES_ROOT
	} else if strings.HasPrefix(p, "HKEY_CURRENT_CONFIG") || strings.HasPrefix(p, "HKCC") {
		k = registry.CURRENT_CONFIG
	}
	if i := strings.IndexRune(s, '\\'); i > 0 {
		p = s[i+1:]
		h, err := registry.OpenKey(k, p, registry.QUERY_VALUE)
		if err != nil {
			return nil, false, err
		}
		defer h.Close()
		if len(v) == 0 {
			return nil, true, nil
		}
		b := make([]byte, 256)
		r, t, err := h.GetValue(v, b)
		if err != nil {
			if err == registry.ErrShortBuffer {
				b = make([]byte, r)
				if _, _, err := h.GetValue(v, b); err != nil {
					return nil, false, err
				}
			} else {
				return nil, false, err
			}
		}
		if t == registry.SZ || t == registry.EXPAND_SZ || t == registry.MULTI_SZ {
			f := (*[1 << 29]uint16)(unsafe.Pointer(&b[0]))[:]
			b = []byte(strings.TrimSpace(string(syscall.UTF16ToString(f))))
			r = len(b)
		}
		return bytes.NewReader(b[:r]), true, nil
	}
	return nil, false, ErrInvalidPrefix
}
