//go:build !windows
// +build !windows

package filter

import "github.com/iDigitalFlame/xmt/util/xerr"

// Select will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices if a PID is not set.
func (f Filter) Select() (uint32, error) {
	if f.PID > 0 {
		return f.PID, nil
	}
	return 0, xerr.Sub("only supported on Windows devices", 0xFA)
}

// Token will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) Token(_ uint32) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0xFA)
}

// Handle will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) Handle(_ uint32) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0xFA)
}

// SelectFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process ID will be returned.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices if a PID is not set.
func (f Filter) SelectFunc(_ filter) (uint32, error) {
	if f.PID > 0 {
		return f.PID, nil
	}
	return 0, xerr.Sub("only supported on Windows devices", 0xFA)
}

// TokenFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as a uint32.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) TokenFunc(a uint32, x filter) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0xFA)
}

// HandleFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
//
// This function allows for a filtering function to be passed along that will be
// supplied with the ProcessID, if the process is elevated, the process name
// and process handle. The function supplied should return true if the process
// passes the filter. The function argument may be nil.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) HandleFunc(_ uint32, _ filter) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0xFA)
}
