package device

import (
	"os"
	"os/user"
)

// IPv6 is a compile-time flag that enables (true) or disables (false) support for IPv6-based
// network addresses.
const IPv6 = true

var (
	// UUID is the device specific and session specific identifier.
	UUID = getID()

	// Arch is the local machine's Architecture type.
	Arch = getArch()

	// Version is the local machine's Operating System version information.
	Version = getVersion()

	// Environment is a mapping of environment string names to their string values.
	Environment = getEnv()
)

// Local is the pointer to the local machine instance. This instance is loaded at runtime and is
// used for local data gathering.
var Local = (&local{&Machine{
	ID:       UUID,
	OS:       OS,
	PID:      uint32(os.Getpid()),
	PPID:     uint32(os.Getppid()),
	Arch:     Arch,
	User:     "Unknown",
	Version:  Version,
	Network:  Network{},
	Hostname: "Unknown",
	Elevated: isElevated(),
}}).init()

type local struct {
	*Machine
}

// PID returns the local machine's PID value.
func PID() uint32 {
	return Local.PID
}

// PPID returns the local machine's Parent PID value.
func PPID() uint32 {
	return Local.PPID
}

// User returns the local machine's current running Username.
func User() string {
	return Local.User
}

// Net returns the local machine's Network data.
func Net() Network {
	return Local.Network
}

// Elevated returns the local machine's Elevated running status.
func Elevated() bool {
	return Local.Elevated
}

// Hostname returns the local machine's Hostname value.
func Hostname() string {
	return Local.Hostname
}
func (l *local) init() *local {
	if u, err := user.Current(); err == nil {
		l.User = u.Username
	}
	l.Hostname, _ = os.Hostname()
	l.Network.Refresh()
	return l
}
func (l *local) Refresh() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	l.User = u.Username
	if l.Hostname, err = os.Hostname(); err != nil {
		return err
	}
	if err := Local.Network.Refresh(); err != nil {
		return err
	}
	l.PID = uint32(os.Getpid())
	l.PPID = uint32(os.Getppid())
	l.Elevated = isElevated()
	return nil
}
