//go:build !windows
// +build !windows

package cmd

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/device"
)

// Pid retruns the process ID of the owning process (the process running
// the thread.)
//
// This may return zero if the thread has not yet been started.
func (DLL) Pid() uint32 {
	return 0
}

// Start will attempt to start the DLL and will return an errors that occur while
// starting the DLL.
//
// This function will return 'ErrEmptyCommand' if the 'Path' parameter is empty
// and 'ErrAlreadyStarted' if attempting to start a DLL that already has been
// started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (DLL) Start() error {
	return device.ErrNoWindows
}

// SetParent will instruct the DLL to choose a parent with the supplied process
// Filter. If the Filter is nil this will use the current process (default).
//
// This function has no effect if the device is not running Windows.
func (DLL) SetParent(_ *filter.Filter) {}
