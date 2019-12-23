package device

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/denisbrodbeck/machineid"
	data "github.com/iDigitalFlame/xmt/xmt-data"
)

const (
	// Windows represents the Windows family of Operating Systems.
	Windows deviceOS = 0x0
	// Linux represents the Linux family of Operating Systems
	Linux deviceOS = 0x1
	// Unix represents the Unix family of Operating Systems
	Unix deviceOS = 0x2
	// Mac represents the MacOS/BSD family of Operating Systems
	Mac deviceOS = 0x3

	// Arch64 represents the 64-bit chipset family.
	Arch64 deviceArch = 0x0
	// Arch86 represents the 32-bit chipset family.
	Arch86 deviceArch = 0x1
	// ArchARM represents the ARM chipset family.
	ArchARM deviceArch = 0x2
	// ArchPowerPC represents the PowerPC chipset family.
	ArchPowerPC deviceArch = 0x3
	// ArchMips represents the MIPS chipset family.
	ArchMips deviceArch = 0x4
	// ArchUnknown represents an unknown chipset family.
	ArchUnknown deviceArch = 0x5

	// IDSize is the amount of bytes used to store the Host ID and
	// SessionID values.  The ID is the (HostID + SessionID).
	IDSize = 32
	// SmallIDSize is the amount of bytes used for printing the Host ID
	// value using the ID function.
	SmallIDSize = MachineIDSize
	// MachineIDSize is the amount of bytes that is used as the Host
	// specific ID value that does not change when on the same host.
	MachineIDSize = 28

	xmtID              = "xmtFramework"
	xmtIDPrime  uint32 = 16777619
	xmtIDOffset uint32 = 2166136261
)

var (
	// Environment is a mapping of environment string names
	// to their string values.
	Environment = make(map[string]string)

	// ErrInvalidID is returned by the 'IDFromString' function when
	// the returned ID value is invalid or nil.
	ErrInvalidID = errors.New("id format is invalid or empty")

	envRegexp = regexp.MustCompile("(%([\\w\\d()-_]+)%|\\$([[\\w\\d-_]+))")
)

// ID is an alias for a byte array that represents a 48 byte
// client identification number.  This is used for tracking and
// detection purposes.
type ID []byte
type deviceOS uint8
type deviceArch uint8

func init() {
	for _, v := range os.Environ() {
		if i := strings.IndexRune(v, '='); i > 0 {
			Environment[strings.ToLower(v[:i])] = v[i+1:]
		}
	}
	Environment["tmp"] = os.TempDir()
	Environment["temp"] = Environment["tmp"]
	Environment["tmpdir"] = Environment["tmp"]
	Environment["tempdir"] = Environment["tmp"]
}
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

// Hash returns the 32bit hash sum of this ID value.
// The hash mechanism used is similar to the hash/fnv mechanism.
func (i ID) Hash() uint32 {
	h := xmtIDOffset
	for x := range i {
		h *= xmtIDPrime
		h ^= uint32(i[x])
	}
	return h
}

// String returns a representation of this ID instance.
func (i ID) String() string {
	if len(i) < SmallIDSize {
		return strings.ToUpper(hex.EncodeToString(i))
	}
	return strings.ToUpper(hex.EncodeToString(i[SmallIDSize:]))
}

// Expand attempts to determine environment variables from the current
// session and translate them in the supplied string.
func Expand(s string) string {
	v := envRegexp.FindAllStringSubmatch(s, -1)
	if len(v) > 0 {
		for _, i := range v {
			if len(i) == 4 {
				n := i[2]
				if len(i[3]) > 0 {
					n = i[3]
				}
				if d, ok := Environment[strings.ToLower(n)]; ok {
					s = strings.Replace(s, i[0], d, -1)
				}
			}
		}
	}
	return s
}

// FullString returns the full string representation of this ID instance.
func (i ID) FullString() string {
	return strings.ToUpper(hex.EncodeToString(i))
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

// IDFromString attempts to convert the hex string supplied
// into an ID value.
func IDFromString(s string) (ID, error) {
	i, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	if i == nil || len(i) == 0 {
		return nil, ErrInvalidID
	}
	switch len(i) {
	case 4, MachineIDSize, IDSize:
		break
	default:
		return nil, fmt.Errorf("id size %d is invalid", len(i))
	}
	return ID(i), nil
}

// MarshalStream writes the data of this ID to the supplied Writer.
func (i ID) MarshalStream(w data.Writer) error {
	_, err := w.Write(i)
	return err
}

// UnmarshalStream reads the data of this ID from the supplied Reader.
func (i *ID) UnmarshalStream(r data.Reader) error {
	if *i == nil {
		*i = append(*i, make([]byte, IDSize)...)
	}
	n, err := data.ReadFully(r, *i)
	if err != nil {
		return err
	}
	if n != IDSize {
		return io.EOF
	}
	return nil
}
