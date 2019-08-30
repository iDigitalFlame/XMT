package device

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"runtime"

	"github.com/denisbrodbeck/machineid"
	"github.com/iDigitalFlame/xmt/xmt/data"
)

const (
	// Windows repersents the Windows family of Operating Systems.
	Windows deviceOS = 0x0
	// Linux repersents the Linux family of Operating Systems
	Linux deviceOS = 0x1
	// Unix repersents the Unix family of Operating Systems
	Unix deviceOS = 0x2
	// Mac repersents the MacOS/BSD family of Operating Systems
	Mac deviceOS = 0x3

	// Arch64 repersents the 64-bit chipset family.
	Arch64 deviceArch = 0x0
	// Arch86 repersents the 32-bit chipset family.
	Arch86 deviceArch = 0x1
	// ArchARM repersents the ARM chipset family.
	ArchARM deviceArch = 0x2
	// ArchPowerPC repersents the PowerPC chipset family.
	ArchPowerPC deviceArch = 0x3
	// ArchMips repersents the MIPS chipset family.
	ArchMips deviceArch = 0x4
	// ArchUnknown repersents an unknown chipset family.
	ArchUnknown deviceArch = 0x5

	// IDSize is the amount of bytes used to store the Host ID and
	// SessionID values.  The ID is the (HostID + SessionID).
	IDSize = 32

	// SmallIDSize is the amount of bytes used for printing the Host ID
	// value using the ID function.
	SmallIDSize = 8

	// MachineIDSize is the amount of bytes that is used as the Host
	// specific ID value that does not change when on the same host.
	MachineIDSize = 28

	xmtID = "xmtFramework"
)

// ID is an alias for a byte array that reperents a 48 byte
// client identification number.  This is used for tracking and
// detection purposes.
type ID []byte
type deviceOS uint8
type deviceArch uint8

func getID() ID {
	i := ID(make([]byte, IDSize))
	s, err := machineid.ProtectedID(xmtID)
	if err == nil {
		copy(i, s)
	} else {
		rand.Read(i)
	}
	rand.Read(i[MachineIDSize:])
	return i
}

// ID returns a small string repersentation of this ID instance.
func (i ID) ID() string {
	if len(i) < SmallIDSize {
		return i.String()
	}
	return hex.EncodeToString(i[:SmallIDSize])
}
func getArch() deviceArch {
	switch runtime.GOARCH {
	case "386":
		return Arch86
	case "amd64", "amd64p32":
		return Arch64
	case "ppc", "ppc64", "ppc64le":
		return ArchPowerPC
	case "arm", "armbe", "arm64", "arm64be":
		return ArchARM
	case "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le":
		return ArchMips
	}
	return ArchUnknown
}

// String returns a repersentation of this ID instance.
func (i ID) String() string {
	return hex.EncodeToString(i)
}
func (d deviceOS) String() string {
	switch d {
	case Windows:
		return "Windows"
	case Linux:
		return "Linux"
	case Unix:
		return "Unix/BSD"
	case Mac:
		return "MacOS"
	}
	return "Unknown"
}
func (d deviceArch) String() string {
	switch d {
	case Arch86:
		return "32bit"
	case Arch64:
		return "64bit"
	case ArchARM:
		return "ARM"
	case ArchMips:
		return "MIPS"
	case ArchPowerPC:
		return "PowerPC"
	}
	return "Unknown"
}

// MarshalStream writes the data of this ID to the supplied Writer.
func (i *ID) MarshalStream(w data.Writer) error {
	if _, err := w.Write(*i); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream reads the data of this ID from the supplied Reader.
func (i *ID) UnmarshalStream(r data.Reader) error {
	if *i == nil {
		*i = append(*i, make([]byte, IDSize)...)
	}
	n, err := r.Read(*i)
	if err != nil {
		return err
	}
	if n != IDSize {
		return io.EOF
	}
	return nil
}
