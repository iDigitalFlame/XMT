package device

import "github.com/iDigitalFlame/xmt/data"

// Machine is a struct that contains information about a specific device.
// This struct contains generic Operating System Information such as Version,
// Arch and network information.
type Machine struct {
	User     string
	Version  string
	Hostname string

	Network   Network
	PID, PPID uint32

	ID       ID
	Arch     deviceArch
	OS       deviceOS
	Elevated uint8
}

// IsElevated will return true if the elevated flag is set to true on this
// device's 'Elevated' flags.
func (m Machine) IsElevated() bool {
	return m.Elevated&1 != 0
}

// IsDomainJoined will return true if the domain joined flag is set to true
// on this device's 'Elevated' flags.
func (m Machine) IsDomainJoined() bool {
	return m.OS == Windows && m.Elevated&128 != 0
}

// MarshalStream transforms this struct into a binary format and writes to the
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
	if err := w.WriteUint8(m.Elevated); err != nil {
		return err
	}
	if err := m.Network.MarshalStream(w); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream transforms this struct from a binary format that is read from
// the supplied data.Reader.
func (m *Machine) UnmarshalStream(r data.Reader) error {
	if err := m.ID.UnmarshalStream(r); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)((&m.OS))); err != nil {
		return err
	}
	if err := r.ReadUint8((*uint8)((&m.Arch))); err != nil {
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
	if err := r.ReadUint8(&m.Elevated); err != nil {
		return err
	}
	if err := m.Network.UnmarshalStream(r); err != nil {
		return err
	}
	return nil
}
