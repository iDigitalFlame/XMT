package local

// This package is used as a simple shortcut for call
// conventions.
// Most calls to local can also be satisfied by device.Local.
// local.Hostname() is equal to device.Local.Hostname for example.

import "github.com/iDigitalFlame/xmt/device"

// OS returns the local machine's Operating System type.
func OS() uint8 {
	return uint8(device.Local.OS)
}

// Arch returns the local machine's Architecture type.
func Arch() uint8 {
	return uint8(device.Local.Arch)
}

// PID returns the local machine's PID value.
func PID() uint64 {
	return device.Local.PID
}

// User returns the local machine's User value.
func User() string {
	return device.Local.User
}

// ID returns the local machine's ID value.
func ID() device.ID {
	return device.Local.ID
}

// OSName returns the local machine's Operating System name.
func OSName() string {
	return device.Local.OS.String()
}

// Elevated returns the local machine's Elevated value.
func Elevated() bool {
	return device.Local.Elevated
}

// Version returns the local machine's Version value.
func Version() string {
	return device.Local.Version
}

// ArchName returns the local machine's Architecture name.
func ArchName() string {
	return device.Local.Arch.String()
}

// Hostname returns the local machine's Hostname value.
func Hostname() string {
	return device.Local.Hostname
}

// Host returns the pointer to the Local Machine struct.
func Host() *device.Machine {
	return device.Local.Machine
}

// Network returns the local machine's Network data.
func Network() device.Network {
	return device.Local.Network
}

// Machine returns the pointer to the Local Machine struct.
func Machine() *device.Machine {
	return device.Local.Machine
}
