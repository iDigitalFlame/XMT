//go:build js || wasm || plan9
// +build js wasm plan9

package device

import (
	"os/exec"
	"os/user"
	"strings"
)

const (
	// OS is the local machine's Operating System type.
	OS = Unsupported

	// Shell is the default machine specific command shell.
	Shell = ""
	// Newline is the machine specific newline character.
	Newline = "\n"
)

// ShellArgs is the default machine specific command shell arguments to run commands.
var ShellArgs = []string{}

func isElevated() bool {
	return false
}
func getVersion() string {
	return "Unsupported"
}
