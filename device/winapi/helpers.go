//go:build windows

package winapi

import (
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

type lsaString struct {
	// DO NOT REORDER
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}
type privileges struct {
	// DO NOT REORDER
	PrivilegeCount uint32
	Privileges     [1]LUIDAndAttributes
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
// This function should NOT be used on real binary files and only used on
// loaded libraries.
//
// DO NOT EXPECT ANYTHING (INCLUDING DEFERS) TO HAPPEN AFTER THIS FUNCTION.
func KillRuntime() {
	runtime.LockOSThread()
	h, err := CreateToolhelp32Snapshot(0x4, 0)
	if err != nil {
		runtime.UnlockOSThread()
		return
	}
	var (
		m = make(map[uint32][]uintptr, 16)
		p = GetCurrentProcessID()
		y = getCurrentThreadID()
		t ThreadEntry32
		v uintptr
		s uint32
	)
	t.Size = uint32(unsafe.Sizeof(t))
	for err = Thread32First(h, &t); err == nil; err = Thread32Next(h, &t) {
		if t.OwnerProcessID != p || t.ThreadID == y {
			continue
		}
		if v, err = OpenThread(0x61, false, t.ThreadID); err != nil {
			break
		}
		if s, err = getThreadStartAddress(v); err != nil {
			break
		}
		if _, ok := m[s]; !ok {
			m[s] = make([]uintptr, 0, 8)
		}
		m[s] = append(m[s], v)
	}
	if CloseHandle(h); len(m) == 0 || (err != nil && err != ErrNoMoreFiles) {
		for k, v := range m {
			for i := range v {
				CloseHandle(v[i])
			}
			m[k] = nil
			delete(m, k)
		}
		m = nil
		runtime.UnlockOSThread()
		return
	}
	var (
		z uint32
		x int
	)
	for k, v := range m {
		if len(v) > x {
			x, z = len(v), k
		}
	}
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
		runtime.UnlockOSThread()
		return
	}
	runtime.GC()
	debug.FreeOSMemory()
	TerminateThread(CurrentThread, 0)
	// NOTE(dij): Buck Stops here.
	runtime.UnlockOSThread()
}

// ZeroTraceEvent will attempt to zero out the NtTraceEvent function call with
// a NOP.
//
// This will return an error if it fails.
func ZeroTraceEvent() error {
	var (
		o   uint32
		err = VirtualProtect(funcNtTraceEvent.address()+3, 1, 0x40, &o)
	)
	if err != nil {
		return err
	}
	(*(*[1]byte)(unsafe.Pointer(funcNtTraceEvent.address() + 3)))[0] = 0xC3
	return VirtualProtect(funcNtTraceEvent.address()+3, 1, o, &o)
}

// GetDebugPrivilege is a quick helper function that will attempt to grant the
// caller the "SeDebugPrivilege" privilege.
func GetDebugPrivilege() error {
	var (
		t   uintptr
		err = OpenProcessToken(CurrentProcess, 0x200E8, &t)
	)
	if err != nil {
		return err
	}
	var p privileges
	if err = LookupPrivilegeValue("", debugPriv, &p.Privileges[0].Luid); err != nil {
		CloseHandle(t)
		return err
	}
	p.Privileges[0].Attributes, p.PrivilegeCount = 0x2, 1
	err = AdjustTokenPrivileges(t, false, unsafe.Pointer(&p), uint32(unsafe.Sizeof(p)), nil, nil)
	CloseHandle(t)
	return err
}
func getCurrentThreadID() uint32 {
	r, _, _ := syscall.SyscallN(funcGetCurrentThreadID.address())
	return uint32(r)
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

// IsTokenElevated returns true if this token has a High or System privileges.
func IsTokenElevated(h uintptr) bool {
	var (
		e, n uint32
		err  = GetTokenInformation(h, 0x14, (*byte)(unsafe.Pointer(&e)), uint32(unsafe.Sizeof(e)), &n)
	)
	return err == nil && n == uint32(unsafe.Sizeof(e)) && e != 0
}

// GetTokenUser retrieves access token user account information and SID.
func GetTokenUser(h uintptr) (*TokenUser, error) {
	u, err := getTokenInfo(h, 1, 50)
	if err != nil {
		return nil, err
	}
	return (*TokenUser)(u), nil
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
	if r > 0 {
		return "", err
	}
	v := UTF16ToString(u.Buffer[4:n])
	for i := len(v) - 1; i > 0; i++ {
		if v[i] == '\\' {
			return v[i+1:], nil
		}
	}
	return v, nil
}
func getThreadStartAddress(h uintptr) (uint32, error) {
	var (
		i         uint32
		r, _, err = syscall.SyscallN(funcNtQueryInformationThread.address(), h, 0x9, uintptr(unsafe.Pointer(&i)), 8, 0)
	)
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
		if q := strings.IndexByte(x, 61); q <= 0 {
			if xerr.Concat {
				return nil, xerr.Sub(`invalid env value "`+x+`"`, 0xD)
			}
			return nil, xerr.Sub("invalid env value", 0xD)
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
