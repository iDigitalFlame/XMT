//go:build !windows

package device

import (
	"os"
	"unsafe"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/util/xerr"
	"golang.org/x/net/http/httpproxy"
)

// ErrNoWindows is an error that is returned when a non-Windows device attempts
// a Windows specific function.
var ErrNoWindows = xerr.Sub("only supported on Windows devices", 0xFA)

type stringHeader struct {
	Data uintptr
	Len  int
}

// RevertToSelf function terminates the impersonation of a client application.
// Returns an error if no impersonation is being done.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
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

// SetProcessName will attempt to overrite the process name on *nix systems
// by overriting the argv block.
//
// Returns 'ErrNoNix' on Windows devices.
//
// Found here: https://stackoverflow.com/questions/14926020/setting-process-name-as-seen-by-ps-in-go
func SetProcessName(s string) error {
	var (
		v = (*stringHeader)(unsafe.Pointer(&os.Args[0]))
		d = (*[1 << 30]byte)(unsafe.Pointer(v.Data))[:v.Len]
		n = copy(d, s)
	)
	if n < len(d) {
		d[n] = 0
	}
	return nil
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
	return ErrNoWindows
}

// AdjustPrivileges will attempt to enable the supplied Windows privilege values
// on the current process's Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustPrivileges(_ ...string) error {
	return ErrNoWindows
}

// ImpersonatePipeToken will attempt to impersonate the Token used by the Named
// Pipe client.
//
// This function is only usable on Windows with a Server Pipe handle.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func ImpersonatePipeToken(_ uintptr) error {
	return ErrNoWindows
}

// AdjustTokenPrivileges will attempt to enable the supplied Windows privilege
// values on the supplied process Token.
//
// Errors during encoding, lookup or assignment will be returned and not all
// privileges will be assigned, if they occur.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
func AdjustTokenPrivileges(_ uintptr, _ ...string) error {
	return ErrNoWindows
}
