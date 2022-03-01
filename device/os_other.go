//go:build js || wasm
// +build js wasm

package device

const (
	// OS is the local machine's Operating System type.
	OS = Unsupported

	// Arch is the local machine's platform architecture.
	Arch = ArchWASM

	// Shell is the default machine specific command shell.
	Shell = ""
	// ShellArgs is the default machine specific command shell arguments to run
	// commands.
	ShellArgs = ""
	// PowerShell is the path to the PowerShell binary, which is based on the
	// underlying OS type.
	PowerShell = ""
)

// Mounts attempts to get the mount points on the local device.
//
// On Windows devices, this is the drive letters avaliable, otherwise on nix*
// systems, this will be the mount points on the system.
//
// The return result (if no errors occurred) will be a string list of all the
// mount points (or Windows drive letters).
func Mounts() ([]string, error) {
	return nil, nil
}
