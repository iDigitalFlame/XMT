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

package device

import (
	"os"
	"runtime/debug"
	"syscall"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// ErrNoWindows is an error that is returned when a non-Windows device attempts
// a Windows specific function.
var ErrNoWindows = xerr.Sub("only supported on Windows devices", 0x20)

// GoExit attempts to walk through the process threads and will forcefully
// kill all Golang based OS-Threads based on their starting address (which
// should be the same when starting from CGo).
//
// This function can be used on binaries, shared libraries or Zombified processes.
//
// Only works on Windows devices and is a wrapper for 'syscall.Exit(0)' for
// *nix devices.
//
// DO NOT EXPECT ANYTHING (INCLUDING DEFERS) TO HAPPEN AFTER THIS FUNCTION.
func GoExit() {
	syscall.Exit(0)
}

// FreeOSMemory forces a garbage collection followed by an
// attempt to return as much memory to the operating system
// as possible. (Even if this is not called, the runtime gradually
// returns memory to the operating system in a background task.)
//
// On Windows, this function also calls 'SetProcessWorkingSetSizeEx(-1, -1, 0)'
// to force the OS to clear any free'd pages.
func FreeOSMemory() {
	debug.FreeOSMemory()
}
func proxyInit() *config {
	return &config{
		HTTPProxy:  dualEnv("HTTP_PROXY", "http_proxy"),
		HTTPSProxy: dualEnv("HTTPS_PROXY", "https_proxy"),
		NoProxy:    dualEnv("NO_PROXY", "no_proxy"),
		CGI:        os.Getenv("REQUEST_METHOD") != "",
	}
}

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func RevertToSelf() error {
	// TODO(dij): *nix support?
	return ErrNoWindows
}
func dualEnv(o, t string) string {
	if v, ok := syscall.Getenv(o); ok {
		return v
	}
	if v, ok := syscall.Getenv(t); ok {
		return v
	}
	return ""
}

// SetCritical will set the critical flag on the current process. This function
// requires administrative privileges and will attempt to get the
// "SeDebugPrivilege" first before running.
//
// If successful, "critical" processes will BSOD the host when killed or will
// be prevented from running.
//
// The boolean returned is the last Critical status. It's set to True if the
// process was already marked as critical.
//
// Use this function with "false" to disable the critical flag.
//
// NOTE: THIS MUST BE DISABLED ON PROCESS EXIT OTHERWISE THE HOST WILL BSOD!!!
//
// Any errors when setting or obtaining privileges will be returned.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func SetCritical(_ bool) (bool, error) {
	return false, ErrNoWindows
}

// Impersonate attempts to steal the Token in use by the target process of the
// supplied filter.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func Impersonate(_ *filter.Filter) error {
	// TODO(dij): *nix support?
	return ErrNoWindows
}

// ImpersonateUser attempts to log in with the supplied credentials and
// impersonate the logged in account.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// This impersonation is locally based, similar to impersonating a Process token.
//
// This also loads the user profile.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonateUser(_, _, _ string) error {
	// TODO(dij): *nix support?
	return ErrNoWindows
}

// ImpersonateUserNetwork attempts to log in with the supplied credentials and impersonate
// the logged in account.
//
// This will set the permissions of all threads in use by the runtime. Once work
// has completed, it is recommended to call the 'RevertToSelf' function to
// revert the token changes.
//
// This impersonation is network based, unlike impersonating a Process token.
// (Windows-only).
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonateUserNetwork(_, _, _ string) error {
	// TODO(dij): *nix support?
	return ErrNoWindows
}
