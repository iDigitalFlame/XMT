// +build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device/devtools"

// Stop will attempt to terminate the currently running DLL instance.
// Always returns nil on non-Windows devices.
func (*DLL) Stop() error {
	return nil
}

// Start will attempt to start the DLL and will return an errors that occur while starting the DLL.
// This function will return 'ErrEmptyCommand' if the 'Path' parameter is empty or nil and 'ErrAlreadyStarted'
// if attempting to start a DLL that already has been started previously. Always returns 'ErrNoWindows'
// on non-Windows devices.
func (*DLL) Start() error {
	return devtools.ErrNoWindows
}

// SetParent will instruct the Process to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). This function has no effect if the device is not running Windows.
func (*DLL) SetParent(_ string) {}

// SetParentPID will instruct the Process to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Process will choose a parent from a list
// of writable processes. This function has no effect if the device is not running Windows.
func (*DLL) SetParentPID(_ int32) {}

// SetParentRandom will set instruct the Process to choose a parent from the supplied string list on runtime. If this
// list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if the
// device is not running Windows.
func (*DLL) SetParentRandom(_ []string) {}

// SetParentEx will instruct the DLL to choose a parent with the supplied process name. If this string
// is empty, this will use the current process (default). This function has no effect if the device is not running
// Windows.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (*DLL) SetParentEx(_ string, _ bool) {}

// SetParentRandomEx will set instruct the DLL to choose a parent from the supplied string list on runtime.
// If this list is empty or nil, there is no limit to the name of the chosen process. This function has no effect if
// the device is not running Windows.
//
// If the specified bool is true, this function will attempt to choose a high integrity process and will fail if
// none can be opened or found.
func (*DLL) SetParentRandomEx(_ []string, _ bool) {}
