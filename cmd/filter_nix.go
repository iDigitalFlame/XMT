//go:build !windows
// +build !windows

package cmd

import "github.com/iDigitalFlame/xmt/device/devtools"

// Select will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices  if a PID is not set.
func (f Filter) Select() (uint32, error) {
	if f.PID > 0 {
		return f.PID, nil
	}
	return 0, devtools.ErrNoWindows
}

// Handle will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) Handle(_ uint32) (uintptr, error) {
	return 0, devtools.ErrNoWindows
}

// SelectFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices  if a PID is not set.
func (f Filter) SelectFunc(_ filter) (uint32, error) {
	if f.PID > 0 {
		return f.PID, nil
	}
	return 0, devtools.ErrNoWindows
}

// HandleFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the Filter
// are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) HandleFunc(_ uint32, _ filter) (uintptr, error) {
	return 0, devtools.ErrNoWindows
}
