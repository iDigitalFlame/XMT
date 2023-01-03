//go:build !windows

// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

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
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
}

// Token will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as an uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) Token(_ uint32) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
}

// Thread will attempt to find a process with the specified Filter options.
// If a suitable process is found, a handle to the first Thread in the Process
// will be returned. The first argument is the access rights requested, expressed
// as an uint32.
//
// An 'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) Thread(_ uint32) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
}

// Handle will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Handle will be returned.
// The first argument is the access rights requested, expressed as an uint32.
//
// An'ErrNoProcessFound' error will be returned if no processes that match the
// Filter are found.
//
// This function returns 'ErrNoWindows' on non-Windows devices.
func (Filter) Handle(_ uint32) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
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
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
}

// TokenFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, the Process Token Handle will be returned.
// The first argument is the access rights requested, expressed as an uint32.
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
func (Filter) TokenFunc(_ uint32, _ filter) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
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
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
}

// ThreadFunc will attempt to find a process with the specified Filter options.
// If a suitable process is found, a handle to the first Thread in the Process
// will be returned. The first argument is the access rights requested, expressed
// as an uint32.
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
func (Filter) ThreadFunc(_ uint32, _ filter) (uintptr, error) {
	return 0, xerr.Sub("only supported on Windows devices", 0x20)
}
