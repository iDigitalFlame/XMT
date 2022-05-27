//go:build windows

package cmd

import (
	"sort"
	"unsafe"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Processes attempts to gather the current running Processes and returns them
// as a slice of ProcessInfo structs, otherwise any errors are returned.
func Processes() ([]ProcessInfo, error) {
	h, err := winapi.CreateToolhelp32Snapshot(2, 0)
	if err != nil {
		return nil, xerr.Wrap("CreateToolhelp32Snapshot", err)
	}
	var (
		r = make(processList, 0, 64)
		e winapi.ProcessEntry32
		s string
	)
	e.Size = uint32(unsafe.Sizeof(e))
	if err = winapi.GetDebugPrivilege(); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("cmd.Processes(): GetDebugPrivilege failed with err=%s", err)
		}
	}
	for err = winapi.Process32First(h, &e); err == nil; err = winapi.Process32Next(h, &e) {
		if s = winapi.UTF16ToString(e.ExeFile[:]); len(s) == 0 {
			continue
		}
		var u string
		// 0x400 - PROCESS_QUERY_INFORMATION
		if z, err1 := winapi.OpenProcess(0x400, false, e.ProcessID); err1 == nil {
			var t uintptr
			// 0x8 - TOKEN_QUERY
			if err1 = winapi.OpenProcessToken(z, 0x8, &t); err1 == nil {
				u, _ = winapi.UserFromToken(t)
				winapi.CloseHandle(t)
			}
			winapi.CloseHandle(z)
		}
		r = append(r, ProcessInfo{Name: s, User: u, PID: e.ProcessID, PPID: e.ParentProcessID})
	}
	winapi.CloseHandle(h)
	sort.Sort(r)
	return r, nil
}
