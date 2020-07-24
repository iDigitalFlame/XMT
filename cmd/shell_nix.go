// +build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device/devtools"

// ShellExecute calls the Windows ShellExecuteW API function. This will "preform an operation on the specified target"
// from the API documentation. The parameters include the Verb (required), Flags, Working Directory and Arguments.
// The first string specified in args is the value that will fill 'lpFile' and the rest will be filled into the
// 'lpArguments' parameter. Otherwise, if empty, they will both be nil. The error returned will be nil if the function
// call is successful. Always returns 'ErrNoWindows' if the device is not running Windows.
func ShellExecute(_ Verb, _ int32, _ string, _ ...string) error {
	return devtools.ErrNoWindows
}
