//go:build windows
// +build windows

package devtools

import (
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/net/http/httpproxy"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	dllNtdll    = windows.NewLazySystemDLL("ntdll.dll")
	dllWinhttp  = windows.NewLazySystemDLL("winhttp.dll")
	dllKernel32 = windows.NewLazySystemDLL("kernel32.dll")
	dllAdvapi32 = windows.NewLazySystemDLL("advapi32.dll")

	funcIsDebuggerPresent                   = dllKernel32.NewProc("IsDebuggerPresent")
	funcAdjustTokenPrivileges               = dllAdvapi32.NewProc("AdjustTokenPrivileges")
	funcRtlSetProcessIsCritical             = dllNtdll.NewProc("RtlSetProcessIsCritical")
	funcImpersonateNamedPipeClient          = dllAdvapi32.NewProc("ImpersonateNamedPipeClient")
	funcCheckRemoteDebuggerPresent          = dllKernel32.NewProc("CheckRemoteDebuggerPresent")
	funcWinHttpGetDefaultProxyConfiguration = dllWinhttp.NewProc("WinHttpGetDefaultProxyConfiguration")
)

// DO NOT REORDER
type proxyInfo struct {
	AccessType         uint32
	Proxy, ProxyBypass *uint16
}
type privileges struct {
	PrivilegeCount uint32
	Privileges     [5]windows.LUIDAndAttributes
}

// IsDebugged returns true if the current process is attached by a debugger.
func IsDebugged() bool {
	if r, _, _ := funcIsDebuggerPresent.Call(); r > 0 {
		return true
	}
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION, false, windows.GetCurrentProcessId())
	if err != nil {
		return false
	}
	var (
		d       bool
		r, _, _ = funcCheckRemoteDebuggerPresent.Call(uintptr(h), uintptr(unsafe.Pointer(&d)))
	)
	windows.CloseHandle(h)
	return r > 0 && d
}

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done. Always returns 'ErrNoWindows' on non-Windows devices.
func RevertToSelf() error {
	return windows.RevertToSelf()
}
func split(s string) []string {
	if len(s) == 0 {
		return nil
	}
	if len(s) == 1 {
		return []string{s}
	}
	var (
		r []string
		x int
	)
	for i := 1; i < len(s); i++ {
		if s[i] != ';' && s[i] != ' ' {
			continue
		}
		if x == i {
			continue
		}
		for ; x < i && (s[x] == ';' || s[x] == ' '); x++ {
		}
		if x == i {
			continue
		}
		r = append(r, s[x:i])
		if x = i + 1; x > len(s) {
			break
		}
		for ; x < len(s) && (s[x] == ';' || s[x] == ' '); x++ {
		}
		i = x
	}
	if x == 0 && len(r) == 0 {
		return []string{s}
	}
	if x < len(s) {
		r = append(r, s[x:])
	}
	return r
}

// SetCritical will set the critical flag on the current process. This function
// requires administrative privileges and will attempt to get the
// "SeDebugPrivilege" first before running.
//
// If successful, "critical" processes will BSOD the host when killed or will
// be prevented from running.
//
// Use this function with "false" to disable the critical flag.
//
// NOTE: THIS MUST BE DISABED ON PROCESS EXIT OTHERWISE THE HOST WILL BSOD!!!
//
// Any errors when setting or obtaining privileges will be returned.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func SetCritical(c bool) error {
	if err := AdjustPrivileges("SeDebugPrivilege"); err != nil {
		return err
	}
	var s byte
	if c {
		s = 1
	}
	if r, _, err := funcRtlSetProcessIsCritical.Call(uintptr(s), 0, 0); r > 0 {
		return err
	}
	return nil
}
func proxyInit() *httpproxy.Config {
	var (
		i         proxyInfo
		r, _, err = funcWinHttpGetDefaultProxyConfiguration.Call(uintptr(unsafe.Pointer(&i)))
	)
	if r != 1 {
		if bugtrack.Enabled {
			bugtrack.Track("devtools.proxyInit(): Retriving proxy info failed, falling back to no proxy: %s", err)
		}
		return nil
	}
	if i.AccessType < 3 || (i.Proxy == nil && i.ProxyBypass == nil) {
		return nil
	}
	var (
		v = windows.UTF16PtrToString(i.Proxy)
		b = windows.UTF16PtrToString(i.ProxyBypass)
	)
	if len(v) == 0 && len(b) == 0 {
		return nil
	}
	var c httpproxy.Config
	if i := split(b); len(i) > 0 {
		c.NoProxy = strings.Join(i, ",")
	}
	for _, x := range split(v) {
		if len(x) == 0 {
			continue
		}
		q := strings.IndexByte(x, '=')
		if q > 1 {
			switch strings.ToLower(x[0:q]) {
			case "http":
				c.HTTPProxy = x[q+1:]
			case "https":
				c.HTTPSProxy = x[q+1:]
			}
			continue
		}
		if len(c.HTTPProxy) == 0 {
			c.HTTPProxy = x
		}
		if len(c.HTTPSProxy) == 0 {
			c.HTTPSProxy = x
		}
	}
	if bugtrack.Enabled {
		bugtrack.Track(
			"devtools.proxyInit(): Proxy info c.HTTPProxy=%s, c.HTTPSProxy=%s, c.NoProxy=%s",
			c.HTTPProxy, c.HTTPSProxy, c.NoProxy,
		)
	}
	return &c
}

// AdjustPrivileges will attempt to enable the supplied Windows privilege values on the current process's Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustPrivileges(s ...string) error {
	var t windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_WRITE|windows.TOKEN_QUERY, &t); err != nil {
		return xerr.Wrap("cannot get current token", err)
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
		if i > 5 {
			break
		}
		if n, err = windows.UTF16PtrFromString(s[i]); err != nil {
			return xerr.Wrap(`cannot convert "`+s[i]+`"`, err)
		}
		if err = windows.LookupPrivilegeValue(nil, n, &p.Privileges[i].Luid); err != nil {
			return xerr.Wrap(`cannot lookup privilege "`+s[i]+`"`, err)
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
	return xerr.Wrap("cannot assign all privileges", err)
}

// ImpersonatePipeToken will attempt to impersonate the Token used by the Named Pipe client. This function is only
// usable on Windows with a Server Pipe handle. Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonatePipeToken(h uintptr) error {
	if r, _, err := funcImpersonateNamedPipeClient.Call(h); r == 0 {
		return err
	}
	return nil
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
		return nil, xerr.New("registry: path does not contain a valid key root")
	}
	i := strings.IndexByte(key, '\\')
	if i <= 0 {
		return nil, xerr.New("registry: path does not contain a valid key root")
	}
	h, err := registry.OpenKey(k, key[i+1:], registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	var m time.Time
	if z, err2 := h.Stat(); err2 == nil {
		m = z.ModTime()
	}
	if len(value) == 0 {
		h.Close()
		return &RegistryFile{k: key, m: m}, nil
	}
	r, t, err := h.GetValue(value, nil)
	if err != nil {
		h.Close()
		return nil, xerr.Wrap(`registry: unable to read path "`+key+`:`+value+`"`, err)
	}
	if r <= 0 {
		h.Close()
		return nil, xerr.New(`registry: path "` + key + `:` + value + `" returned a zero size`)
	}
	b := make([]byte, r)
	_, _, err = h.GetValue(value, b)
	if h.Close(); err != nil {
		return nil, xerr.Wrap(`registry: unable to read path "`+key+`:`+value+`"`, err)
	}
	return &RegistryFile{k: key, v: value, t: byte(t), m: m, b: b}, nil
}

// AdjustTokenPrivileges will attempt to enable the supplied Windows privilege values on the supplied process Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustTokenPrivileges(h uintptr, s ...string) error {
	if len(s) <= 5 {
		return adjust(h, s)
	}
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
