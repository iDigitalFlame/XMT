// +build windows

package cmd

import (
	"strings"

	"golang.org/x/sys/windows"
)

// ShellExecute calls the Windows ShellExecuteW API function. This will "preform an operation on the specified target"
// from the API documentation. The parameters include the Verb (required), Flags, Working Directory and Arguments.
// The first string specified in args is the value that will fill 'lpFile' and the rest will be filled into the
// 'lpArguments' parameter. Otherwise, if empty, they will both be nil. The error returned will be nil if the function
// call is successful. Always returns 'ErrNoWindows' if the device is not running Windows.
func ShellExecute(v Verb, f int32, dir string, args ...string) error {
	var (
		err        error
		o, e, d, a *uint16
	)
	if o, err = windows.UTF16PtrFromString(string(v)); err != nil {
		return err
	}
	if len(dir) > 0 {
		if d, err = windows.UTF16PtrFromString(dir); err != nil {
			return err
		}
	}
	if len(args) > 0 {
		if e, err = windows.UTF16PtrFromString(args[0]); err != nil {
			return err
		}
		if len(args) > 1 {
			if a, err = windows.UTF16PtrFromString(strings.Join(args, " ")); err != nil {
				return err
			}
		}
	}
	return windows.ShellExecute(0, o, e, a, d, f)
}
