//go:build windows && go1.11
// +build windows,go1.11

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

package winapi

import (
	"strings"
	"sync"
	"unsafe"
)

const (
	tokenPerms   = 0
	fallbackLoad = false
)

var dllKernelOrAdvapi = dllKernelBase

var compatOnce struct {
	sync.Once
	v bool
}

// IsWindows7 returns true if the underlying device runs at least Windows 7
// (>=6.2) and built using <= go1.10.
//
// If built using >= go1.11, this function always returns true.
func IsWindows7() bool {
	return true
}

// EmptyWorkingSet Windows API Call wrapper
//
//	Removes as many pages as possible from the working set of the specified
//	process.
//
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-emptyworkingset
//
// Wraps the 'SetProcessWorkingSetSizeEx' call instead to prevent having to track
// the 'EmptyWorkingSet' function between kernel32.dll and psapi.dll.
//
// This function will fallback to 'SetProcessWorkingSetSize' if the underlying
// device is Windows Xp and built using <= go1.10.
func EmptyWorkingSet() {
	syscallN(funcSetProcessWorkingSetSizeEx.address(), CurrentProcess, invalid, invalid, 0)
}
func checkCompatFunc() {
	if m, _, _ := GetVersionNumbers(); m >= 10 {
		compatOnce.v = true
	}
}

// IsWindowsXp returns true if the underlying device is Windows Xp and NOT Server
// 2003.
//
// If built using >= go1.11, this function always returns false.
func IsWindowsXp() bool {
	return false
}

// IsWindows10 returns true if the underlying device runs at least Windows 10
// (>=10).
func IsWindows10() bool {
	compatOnce.Do(checkCompatFunc)
	return compatOnce.v
}

// IsWindowsVista returns true if the underlying device runs at least Windows Vista
// (>=6) and built using <= go1.10.
//
// If built using >= go1.11, this function always returns true.
func IsWindowsVista() bool {
	return true
}

// UserInAdminGroup returns true if the current thread or process token user is
// part of the Administrators group. This is only used if the device is older than
// Windows Vista and built using <= go1.10.
//
// If built using >= go1.11, this function always returns false.
func UserInAdminGroup() bool {
	return false
}

// IsTokenElevated returns true if this token has a High or System privileges.
//
// Always returns false on any systems older than Windows Vista.
func IsTokenElevated(h uintptr) bool {
	var (
		e, n uint32
		err  = GetTokenInformation(h, 0x14, (*byte)(unsafe.Pointer(&e)), 4, &n)
		// 0x14 - TokenElevation
	)
	return err == nil && n == 4 && e != 0
}

// IsWow64Process Windows API Call
//
//	Determines whether the specified process is running under WOW64 or an
//	Intel64 of x64 processor.
//
// https://docs.microsoft.com/en-us/windows/win32/api/wow64apiset/nf-wow64apiset-iswow64process
func IsWow64Process(h uintptr) (bool, error) {
	if funcRtlWow64GetProcessMachines.Load() != nil {
		// Running on "true" x86.
		return false, nil
	}
	var p, n uint16
	if r, _, _ := syscallN(funcRtlWow64GetProcessMachines.address(), h, uintptr(unsafe.Pointer(&p)), uintptr(unsafe.Pointer(&n))); r > 0 {
		return false, formatNtError(r)
	}
	return p > 0, nil
}

// CancelIoEx Windows API Call
//
//	Marks any outstanding I/O operations for the specified file handle. The
//	function only cancels I/O operations in the current process, regardless of
//	which thread created the I/O operation.
//
// https://docs.microsoft.com/en-us/windows/win32/fileio/cancelioex-func
//
// Re-targeted to use 'NtCancelIoFileEx' instead.
// https://learn.microsoft.com/en-us/windows/win32/devnotes/nt-cancel-io-file-ex
//
// NOTE(dij): ^ THIS IS WRONG! It forgets the IO_STATUS_BLOCK entry at the end.
//
//	NtCancelIoFileEx (HANDLE FileHandle, PIO_STATUS_BLOCK IoRequestToCancel, PIO_STATUS_BLOCK IoStatusBlock)
//
// This function will fallback to 'NtCancelIoFile' if the underlying device is
// older than Windows 7 and built using <= go1.10.
//
// Normally, Windows Vista would work, but this has a weird issue that causes
// it to wait forever.
func CancelIoEx(h uintptr, o *Overlapped) error {
	var s [4 + ptrSize]byte // IO_STATUS_BLOCK
	if r, _, _ := syscallN(funcNtCancelIoFileEx.address(), h, uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(&s))); r > 0 {
		return formatNtError(r)
	}
	return nil
}
func copyMemory(d uintptr, s uintptr, x uint32) {
	syscallN(funcRtlCopyMappedMemory.address(), d, s, uintptr(x))
}

// RegDeleteTree Windows API Call
//
//	Deletes the subkeys and values of the specified key recursively.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winreg/nf-winreg-regdeletetreew
//
// This function returns 'syscall.EINVAL' if the underlying device is older than
// Windows Vista and built using <= go1.10.
func RegDeleteTree(h uintptr, path string) error {
	p, err := UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, err1 := syscallN(funcRegDeleteTree.address(), h, uintptr(unsafe.Pointer(p)))
	if r > 0 {
		return unboxError(err1)
	}
	return nil
}

// EnumDrivers attempts to reterive the list of currently loaded drivers
// and will call the supplied function with the handle of each driver along with
// the base name of the driver file.
//
// The user supplied function can return an error that if non-nil, will stop
// Driver iteration immediately and will be returned by this function.
//
// Callers can return the special 'winapi.ErrNoMoreFiles' error that will stop
// iteration but will cause this function to return nil. This can be used to
// stop iteration without errors if needed.
func EnumDrivers(f func(uintptr, string) error) error {
	var (
		n          uint32
		r, _, err1 = syscallN(funcK32EnumDeviceDrivers.address(), 0, 0, uintptr(unsafe.Pointer(&n)))
	)
	if r == 0 {
		return unboxError(err1)
	}
	e := make([]uintptr, (n/uint32(ptrSize))+32)
	r, _, err1 = syscallN(funcK32EnumDeviceDrivers.address(), uintptr(unsafe.Pointer(&e[0])), uintptr(n+uint32(32*ptrSize)), uintptr(unsafe.Pointer(&n)))
	if r == 0 {
		return unboxError(err1)
	}
	var (
		s   [260]uint16
		err error
		b   = UTF16ToString((*kernelSharedData)(unsafe.Pointer(kernelShared)).NtSystemRoot[:])
	)
	for i := range e {
		if e[i] == 0 {
			continue
		}
		if r, _, err1 = syscallN(funcK32GetDeviceDriverFileName.address(), e[i], uintptr(unsafe.Pointer(&s[0])), 260); r == 0 {
			return unboxError(err1)
		}
		v := strings.Replace(UTF16ToString(s[:r]), sysRoot, b, 1)
		if len(v) > 5 && v[0] == '\\' && v[1] == '?' && v[3] == '\\' {
			v = v[4:]
		}
		if err = f(e[i], v); err != nil {
			break
		}
	}
	if err != nil && err == ErrNoMoreFiles {
		return err
	}
	return nil
}
func getCurrentModuleInfo(h uintptr, i *modInfo) error {
	if r, _, err := syscallN(funcK32GetModuleInformation.address(), CurrentProcess, h, uintptr(unsafe.Pointer(i)), ptrSize*3); r == 0 {
		return err
	}
	return nil
}

// RegDeleteKeyEx Windows API Call
//
//	Deletes a subkey and its values. Note that key names are not case sensitive.
//	ONLY DELETES EMPTY SUBKEYS. (invalid argument if non-empty)
//
// https://docs.microsoft.com/en-us/windows/win32/api/winreg/nf-winreg-regdeletekeyexw
//
// This function will fallback to 'RegDeleteKey' if the underlying device is
// older than Windows Vista and built using <= go1.10.
func RegDeleteKeyEx(h uintptr, path string, f uint32) error {
	p, err := UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	r, _, err1 := syscallN(funcRegDeleteKeyEx.address(), h, uintptr(unsafe.Pointer(p)), uintptr(f), 0)
	if r > 0 {
		return unboxError(err1)
	}
	return nil
}

// WinHTTPGetDefaultProxyConfiguration Windows API Call
//
//	The WinHttpGetDefaultProxyConfiguration function retrieves the default WinHTTP
//	proxy configuration from the registry.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winhttp/nf-winhttp-winhttpgetdefaultproxyconfiguration
//
// This function returns 'syscall.EINVAL' if the underlying device is Windows Xp
// and built using <= go1.10.
func WinHTTPGetDefaultProxyConfiguration(i *ProxyInfo) error {
	r, _, err := syscallN(funcWinHTTPGetDefaultProxyConfiguration.address(), uintptr(unsafe.Pointer(&i)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// NtCreateThreadEx Windows API Call
//
//	Creates a thread that runs in the virtual address space of another process
//	and optionally specifies extended attributes such as processor group affinity.
//
// http://pinvoke.net/default.aspx/ntdll/NtCreateThreadEx.html
//
// This function will fallback to 'CreateRemoteThread' if the underlying device
// is older than Windows Vista and built using <= go1.10.
func NtCreateThreadEx(h, address, args uintptr, suspended bool) (uintptr, error) {
	// TODO(dij): Add additional injection types?
	//            - NtQueueApcThread
	//            - Kernel Table Callback
	f := uint32(0x4)
	// 0x4 - THREAD_CREATE_FLAGS_HIDE_FROM_DEBUGGER
	if suspended {
		// 0x1 - CREATE_SUSPENDED
		f |= 0x1
	}
	var (
		t       uintptr
		r, _, _ = syscallN(
			funcNtCreateThreadEx.address(), uintptr(unsafe.Pointer(&t)), 0x10000000, 0, h, address, args, uintptr(f),
			0, 0, 0, 0,
		)
		// 0x10000000 - THREAD_ALL_ACCESS
	)
	if r > 0 {
		return 0, formatNtError(r)
	}
	return t, nil
}

// CreateProcessWithToken Windows API Call
//
//	Creates a new process and its primary thread. The new process runs in the
//	security context of the specified token. It can optionally load the user
//	profile for the specified user.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-createprocesswithtokenw
//
// This function returns 'syscall.EINVAL' if the underlying device is Windows Xp
// and built using <= go1.10.
func CreateProcessWithToken(t uintptr, loginFlags uint32, name, cmd string, flags uint32, env []string, dir string, y *StartupInfo, x *StartupInfoEx, i *ProcessInformation) error {
	var (
		n, c, d, e *uint16
		err        error
	)
	if len(name) > 0 {
		if n, err = UTF16PtrFromString(name); err != nil {
			return err
		}
	}
	if len(cmd) > 0 {
		if c, err = UTF16PtrFromString(cmd); err != nil {
			return err
		}
	}
	if len(dir) > 0 {
		if d, err = UTF16PtrFromString(dir); err != nil {
			return err
		}
	}
	if len(env) > 0 {
		if e, err = StringListToUTF16Block(env); err != nil {
			return err
		}
		// 0x400 - CREATE_UNICODE_ENVIRONMENT
		flags |= 0x400
	}
	var j unsafe.Pointer
	if y == nil && x != nil {
		// BUG(dij): For some reason adding this flag causes the function
		//           to return "invalid parameter", even this this IS THE ACCEPTED
		//           thing to do???!
		//
		// flags |= 0x80000
		j = unsafe.Pointer(x)
	} else {
		j = unsafe.Pointer(y)
	}
	r, _, err1 := syscallN(
		funcCreateProcessWithToken.address(), t, uintptr(loginFlags), uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(c)),
		uintptr(flags), uintptr(unsafe.Pointer(e)), uintptr(unsafe.Pointer(d)), uintptr(j), uintptr(unsafe.Pointer(i)),
	)
	if r == 0 {
		return unboxError(err1)
	}
	return nil
}
