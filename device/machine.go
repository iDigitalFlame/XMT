package device

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
)

// Machine is a struct that contains information about a specific device. This struct contains generic Operating
// System Information such as Version, Arch and network information.
type Machine struct {
	User string `json:"user"`

	Version  string `json:"version"`
	Hostname string `json:"hostname"`

	ID      ID      `json:"id"`
	Network Network `json:"network"`

	PID  uint32 `json:"pid"`
	PPID uint32 `json:"ppid"`

	Arch     deviceArch `json:"arch"`
	OS       deviceOS   `json:"os"`
	Elevated bool       `json:"elevated"`
}

// String returns a simple string representation of the Machine instance.
func (m Machine) String() string {
	var e string
	if m.Elevated {
		e = "*"
	}
	return "[" + m.ID.String() + "] " + m.Hostname + " (" + m.Version + ") " + e + m.User
}

// MarshalStream transforms this struct into a binary format and writes to the supplied data.Writer.
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
	if err := w.WriteUint32(m.PID); err != nil {
		return err
	}
	if err := w.WriteUint32(m.PPID); err != nil {
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

// UnmarshalStream transforms this struct from a binary format that is read from the supplied data.Reader.
func (m *Machine) UnmarshalStream(r data.Reader) error {
	if err := m.ID.UnmarshalStream(r); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)(unsafe.Pointer(&m.OS))); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)(unsafe.Pointer(&m.Arch))); err != nil {
		return err
	}
	if err := r.ReadUint32(&m.PID); err != nil {
		return err
	}
	if err := r.ReadUint32(&m.PPID); err != nil {
		return err
	}
	if err := r.ReadString(&m.User); err != nil {
		return err
	}
	if err := r.ReadString(&m.Version); err != nil {
		return err
	}
	if err := r.ReadString(&m.Hostname); err != nil {
		return err
	}
	if err := r.ReadBool(&m.Elevated); err != nil {
		return err
	}
	if err := m.Network.UnmarshalStream(r); err != nil {
		return err
	}
	return nil
}
