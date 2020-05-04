// +build !windows

package cmd

type base struct{}

// Stop will attempt to terminate the currently running Code thread instance.
// Always returns nil on non-Windows devices.
func (*Code) Stop() error {
	return nil
}

// Wait will block until the Code thread completes or is terminated by a call to Stop. This function will return
// 'ErrNotCompleted' if the Process has not been started. Always returns nil if the device is not running Windows.
func (*Code) Wait() error {
	return nil
}

// Start will attempt to start the Code thread and will return an errors that occur while starting the Code thread.
// This function will return 'ErrEmptyCommand' if the 'Data' parameter is empty or nil and 'ErrAlreadyStarted'
// if attempting to start a Code thread that already has been started previously. Always returns 'ErrNotSupportedOS'
// on non-Windows devices.
func (*Code) Start() error {
	return ErrNotSupportedOS
}
func (base) String() string {
	return ""
}

// SetParent will instruct the Code thread to choose a parent with the supplied process name. If this string is empty,
// this will use the current process (default). Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Code) SetParent(_ string) error {
	return ErrNotSupportedOS
}

// SetParentPID will instruct the Code thread to choose a parent with the supplied process ID. If this number is
// zero, this will use the current process (default) and if < 0 this Code thread will choose a parent from a list
// of writable processes. Always returns 'ErrNotSupportedOS' if the device is not running Windows.
func (*Code) SetParentPID(_ int32) error {
	return ErrNotSupportedOS
}

// SetParentRandom will set instruct the Code thread to choose a parent from the supplied string list on runtime. If
// this list is empty or nil, there is no limit to the name of the chosen process. Always returns 'ErrNotSupportedOS' if
// the device is not running Windows.
func (*Code) SetParentRandom(_ []string) error {
	return ErrNotSupportedOS
}
