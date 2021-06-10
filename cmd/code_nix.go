// +build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device/devtools"

type base struct{}

// Stop will attempt to terminate the currently running Code thread instance.
// Always returns nil on non-Windows devices.
func (*Code) Stop() error {
	return nil
}

// Start will attempt to start the Code thread and will return an errors that occur while starting the Code thread.
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or nil and 'ErrAlreadyStarted'
// if attempting to start a Code thread that already has been started previously. Always returns 'ErrNoWindows'
// on non-Windows devices.
func (*Code) Start() error {
	return devtools.ErrNoWindows
}
func (base) String() string {
	return ""
}

// SetParent will instruct the Code thread to choose a parent with the supplied process Filter. If the Filter is nil
// this will use the current process (default). This function has no effect if the device is not running Windows.
func (*Code) SetParent(_ *Filter) {}
