//go:build windows

package device

// OS is the local machine's Operating System type.
const OS = Windows

// Shell is the default machine specific command shell.
var Shell = shell()
