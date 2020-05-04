// +build windows

package cmd

import (
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	dllShell32 = windows.NewLazySystemDLL("shell32.dll")

	funcShellExecute = dllShell32.NewProc("ShellExecuteW")
)

// ShellExecute calls the Windows ShellExecuteW API function. This will "preform an operation on the specified target"
// from the API documentation. The parameters include the Verb (required), Flags, Working Directory and Arguments.
// The first string specified in args is the value that will fill 'lpFile' and the rest will be filled into the
// 'lpArguments' parameter. Otherwise, if empty, they will both be nil. The error returned will be nil if the function
// call is successful.
func ShellExecute(v Verb, f int32, dir string, args ...string) error {
	var (
		err        error
		o, e, d, a *uint16
	)
	if o, err = syscall.UTF16PtrFromString(string(v)); err != nil {
		return err
	}
	if len(dir) > 0 {
		if d, err = syscall.UTF16PtrFromString(dir); err != nil {
			return err
		}
	}
	if len(args) > 0 {
		if e, err = syscall.UTF16PtrFromString(args[0]); err != nil {
			return err
		}
		if len(args) > 1 {
			if a, err = syscall.UTF16PtrFromString(strings.Join(args[1:], " ")); err != nil {
				return err
			}
		}
	}
	r, _, err := funcShellExecute.Call(
		uintptr(0),
		uintptr(unsafe.Pointer(o)),
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(a)),
		uintptr(unsafe.Pointer(d)),
		uintptr(f),
	)
	if r != 0 && r <= 32 {
		return err
	}
	return nil
}
