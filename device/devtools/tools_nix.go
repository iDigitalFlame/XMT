//go:build !windows
// +build !windows

package devtools

import (
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/net/http/httpproxy"
)

// ErrNoWindows is an error that is returned when a non-Windows device attempts a Windows specific function.
var ErrNoWindows = xerr.New("only supported on Windows devices")

// IsDebugged returns true if the current process is attached by a debugger.
func IsDebugged() bool {
	return false
}

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done. Always returns 'ErrNoWindows' on non-Windows devices.
func RevertToSelf() error {
	return ErrNoWindows
}

// SetCritical will set the critical flag on the current process. This function
// requires administrative privileges and will attempt to get the
// "SeDebugPrivilege" first before running.
//
// If successful, "critical" processes will BSOD the host when killed or will
// be prevented from running.
//
// Use this function with "false" to disable the critical flag.
//
// NOTE: THIS MUST BE DISABED ON PROCESS EXIT OTHERWISE THE HOST WILL BSOD!!!
//
// Any errors when setting or obtaining privileges will be returned.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func SetCritical(_ bool) error {
	return ErrNoWindows
}
func proxyInit() *httpproxy.Config {
	return httpproxy.FromEnvironment()
}

// AdjustPrivileges will attempt to enable the supplied Windows privilege values on the current process's Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustPrivileges(_ ...string) error {
	return ErrNoWindows
}

// ImpersonatePipeToken will attempt to impersonate the Token used by the Named Pipe client. This function is only
// usable on Windows with a Server Pipe handle. Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonatePipeToken(_ uintptr) error {
	return ErrNoWindows
}

// Registry attempts to open a registry value or key, value pair on Windows devices. Returns err if the system is
// not a Windows device or an error occurred during the open. Always returns 'ErrNoWindows' on non-windows devices.
func Registry(_, _ string) (*RegistryFile, error) {
	return nil, ErrNoWindows
}

// AdjustTokenPrivileges will attempt to enable the supplied Windows privilege values on the supplied process Token.
// Errors during encoding, lookup or assignment will be returned and not all privileges will be assigned, if they
// occur. Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustTokenPrivileges(_ uintptr, _ ...string) error {
	return ErrNoWindows
}
