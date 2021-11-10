//go:build windows
// +build windows

package cmd

import (
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/sys/windows"
)

// Fork will attempt to use built-in system utilities to fork off the process
// into a separate, but similar process. If successful, this function will
// return the PID of the new process.
//
// NOTE(dij): Might be broken :(
//            I think WSL is the reason, Win 7/8 seems "workable".
func Fork() (uint32, error) {
	var (
		i         processInfo
		r, _, err = funcRtlCloneUserProcess.Call(0x0001|0x0002, 0, 0, 0, uintptr(unsafe.Pointer(&i)))
	)
	switch r {
	case 0:
		h, err2 := windows.OpenThread(0x000F|0x00100000|0xFFFF, false, uint32(i.ClientID.Thread))
		if err2 != nil {
			return 0, xerr.Wrap("OpenThread", err2)
		}
		if _, err = windows.ResumeThread(h); err != nil {
			return 0, xerr.Wrap("ResumeThread", err)
		}
		return uint32(i.ClientID.Process), windows.CloseHandle(h)
	case 297:
		if r, _, err = funcAllocConsole.Call(); r == 0 {
			return 0, xerr.Wrap("AllocConsole", err)
		}
		return 0, nil
	}
	return 0, xerr.Wrap("RtlCloneUserProcess", err)
}

// ShellExecute calls the Windows ShellExecuteW API function. This will
// "preform an operation on the specified target" from the API documentation.
//
// The parameters include the Verb (required), Flags, Working Directory and Arguments.
// The first string specified in args is the value that will fill 'lpFile' and the rest
// will be filled into the 'lpArguments' parameter. Otherwise, if empty, they will both be nil.
//
// The error returned will be nil if the function call is successful.
//
// Always returns 'ErrNoWindows' if the device is not running Windows.
func ShellExecute(v Verb, f int32, dir string, args ...string) error {
	var (
		o, _    = windows.UTF16PtrFromString(string(v))
		err     error
		e, d, a *uint16
	)
	if len(dir) > 0 {
		if d, err = windows.UTF16PtrFromString(dir); err != nil {
			return xerr.Wrap("cannot convert dir", err)
		}
	}
	if len(args) > 0 {
		if e, err = windows.UTF16PtrFromString(args[0]); err != nil {
			return xerr.Wrap("cannot convert args", err)
		}
		if len(args) > 1 {
			if a, err = windows.UTF16PtrFromString(strings.Join(args, " ")); err != nil {
				return xerr.Wrap("cannot convert args", err)
			}
		}
	}
	return windows.ShellExecute(0, o, e, a, d, f)
}
