package device

import (
	"os"
	"os/user"

	data "github.com/iDigitalFlame/xmt/xmt-data"
	compat "github.com/iDigitalFlame/xmt/xmt-device/compat"
)

var (
	// Shell is the default machine specific command shell.
	Shell = compat.Shell()

	// Local is the pointer to the local
	// machine instance. This instance is loaded at
	// runtime and is used for local data gathering.
	Local = (&localMachine{
		&Machine{
			ID:       getID(),
			OS:       deviceOS(compat.Os()),
			PID:      uint64(os.Getpid()),
			Arch:     getArch(),
			User:     "Unknown",
			Version:  compat.Version(),
			Network:  Network{},
			Hostname: "Unknown",
			Elevated: compat.Elevated(),
		},
	}).init()

	// Newline is the machine specific newline character.
	Newline = compat.Newline()

	// ShellArgs is the default machine specific command shell
	// arguments to run commands.
	ShellArgs = compat.ShellArgs()
)

// Machine is a struct that contains information about a specific device.
// This struct contains generic Operating System Information such as Version, Arch and
// network information.
type Machine struct {
	ID       ID         `json:"id"`
	OS       deviceOS   `json:"os"`
	PID      uint64     `json:"pid"`
	PPID     uint64     `json:"ppid"`
	Arch     deviceArch `json:"arch"`
	User     string     `json:"user"`
	Version  string     `json:"version"`
	Network  Network    `json:"network"`
	Hostname string     `json:"hostname"`
	Elevated bool       `json:"elevated"`
}
type localMachine struct {
	*Machine
}

func (l *localMachine) Refresh() error {
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
	l.PID = uint64(os.Getpid())
	l.PPID = uint64(os.Getppid())
	l.Elevated = compat.Elevated()
	return nil
}
func (l *localMachine) init() *localMachine {
	if h, err := os.Hostname(); err == nil {
		l.Hostname = h
	}
	if u, err := user.Current(); err == nil {
		l.User = u.Username
	}
	l.Network.Refresh()
	return l
}

// MarshalStream transform this struct into a binary format and writes to the
// supplied data.Writer.
func (m Machine) MarshalStream(w data.Writer) error {
	if err := m.ID.MarshalStream(w); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(m.OS)); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(m.Arch)); err != nil {
		return err
	}
	if err := w.WriteUint64(m.PID); err != nil {
		return err
	}
	if err := w.WriteUint64(m.PPID); err != nil {
		return err
	}
	if err := w.WriteString(m.User); err != nil {
		return err
	}
	if err := w.WriteString(m.Version); err != nil {
		return err
	}
	if err := w.WriteString(m.Hostname); err != nil {
		return err
	}
	if err := w.WriteBool(m.Elevated); err != nil {
		return err
	}
	if err := m.Network.MarshalStream(w); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream transforms this struct from a binary format that is read
// from the supplied data.Reader.
func (m *Machine) UnmarshalStream(r data.Reader) error {
	if err := m.ID.UnmarshalStream(r); err != nil {
		return err
	}
	var err error
	var o, a uint8
	if o, err = r.Uint8(); err != nil {
		return err
	}
	if a, err = r.Uint8(); err != nil {
		return err
	}
	m.OS = deviceOS(o)
	m.Arch = deviceArch(a)
	if err := r.ReadUint64(&(m.PID)); err != nil {
		return err
	}
	if err := r.ReadUint64(&(m.PPID)); err != nil {
		return err
	}
	if err := r.ReadString(&(m.User)); err != nil {
		return err
	}
	if err := r.ReadString(&(m.Version)); err != nil {
		return err
	}
	if err := r.ReadString(&(m.Hostname)); err != nil {
		return err
	}
	if err := r.ReadBool(&(m.Elevated)); err != nil {
		return err
	}
	if err := m.Network.UnmarshalStream(r); err != nil {
		return err
	}
	return nil
}
