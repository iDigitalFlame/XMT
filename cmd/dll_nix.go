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

// SetParent will instruct the DLL to choose a parent with the supplied process Filter. If the Filter is nil
// this will use the current process (default). This function has no effect if the device is not running Windows.
func (*DLL) SetParent(_ *Filter) {}
