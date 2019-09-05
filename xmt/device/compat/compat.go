package compat

import "errors"

var (
	// ErrNoRegistry is an error that is returned when attempting to access the registry
	// on a machine that is not running Windows.
	ErrNoRegistry = errors.New("registry support is only avaliable on Windows")
	// ErrInvalidPrefix is an error returned when attempting to get registry data using
	// an invalid Registry top level key prefix.
	ErrInvalidPrefix = errors.New("cannot find the specified key prefix")
)

// Os returns the Operating system value.
// This is only needed for the Local Device struct, use the
// device or local packages to gain this information in a more
// efficient way.
func Os() uint8 {
	return osv
}

// Shell returns the machine specific default shell value.
// This is only needed for the Local Device struct, use the
// device or local packages to gain this information in a more
// efficient way.
func Shell() string {
	return shell
}

// Elevated returns the user's privilege level as Admin or User.
// This is only needed for the Local Device struct, use the
// device or local packages to gain this information in a more
// efficient way.
func Elevated() bool {
	return getElevated()
}

// Version returns the machine specific Operating System version.
// This is only needed for the Local Device struct, use the
// device or local packages to gain this information in a more
// efficient way.
func Version() string {
	return getVersion()
}

// Newline returns the machine specific newline value.
// This is only needed for the Local Device struct, use the
// device or local packages to gain this information in a more
// efficient way.
func Newline() string {
	return newline
}

// ShellArgs returns the machine specific default shell arguments.
// This is only needed for the Local Device struct, use the
// device or local packages to gain this information in a more
// efficient way.
func ShellArgs() []string {
	return args
}
