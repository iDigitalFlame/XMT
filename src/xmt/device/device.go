package device

import (
	"net"
	"os"
	"runtime"
	"strings"

	"../../xmt/dio"

	"github.com/denisbrodbeck/machineid"
)

const (
	// IPv6 determines if IPv6 support is enabled.
	IPv6 bool = true
	// Windows repersents the Windows family of Operating Systems.
	Windows deviceOS = iota
	// Linux repersents the Linux family of Operating Systems
	Linux
	// Unix repersents the Unix family of Operating Systems
	Unix
	// Mac repersents the MacOS/BSD family of Operating Systems
	Mac

	// Arch64 repersents the 64-bit chipset family.
	Arch64 deviceArch = 0
	// Arch86 repersents the 32-bit chipset family.
	Arch86 deviceArch = iota
	// ArchARM repersents the ARM chipset family.
	ArchARM
	// ArchUnknown repersents an unknown chipset family.
	ArchUnknown
)

var (
	// Local is the repersentation of the current device information.
	Local = getDevice()
	// Enviorment is a string map of current on host enviorment variables.
	Enviorment = map[string]string{}
)

// Address is an alias for a string repersentation of an IP Address.
type Address string
type deviceOS uint8
type deviceArch uint8

// Device is a struct that contains information about a specific device.
// This struct contains generic Operating System Information such as Family, Arch and
// network information.
type Device struct {
	ID       string
	OS       deviceOS
	IPs      []Address
	Arch     deviceArch
	Admin    bool
	Family   string
	Version  string
	Hostname string
}

func init() {
	for _, v := range os.Environ() {
		if i := strings.IndexRune(v, '='); i > 0 {
			Enviorment[strings.ToLower(v[:i])] = v[i+1:]
		}
	}
	Enviorment["tmpdir"] = os.TempDir()
}
func getDevice() *Device {
	d := &Device{}
	if h, err := os.Hostname(); err == nil {
		d.Hostname = h
	}
	if i, err := machineid.ID(); err == nil {
		d.ID = i
	}
	switch runtime.GOARCH {
	case "amd64":
		d.Arch = Arch64
	case "386":
		d.Arch = Arch86
	case "arm":
		d.Arch = ArchARM
	default:
		d.Arch = ArchUnknown
	}
	d.IPs = getDeviceIPs()
	return d
}
func getDeviceIPs() []Address {
	l := make([]Address, 0)
	if i, err := net.Interfaces(); err == nil {
		for _, a := range i {
			if a.Flags&net.FlagUp == 0 || a.Flags&net.FlagLoopback != 0 {
				continue
			}
			if n, err := a.Addrs(); err == nil {
				for _, ad := range n {
					var r net.IP
					switch ad.(type) {
					case *net.IPNet:
						r = ad.(*net.IPNet).IP
					case *net.IPAddr:
						r = ad.(*net.IPAddr).IP
					default:
						continue
					}
					if r.IsLoopback() || r.IsUnspecified() || r.IsMulticast() || r.IsInterfaceLocalMulticast() || r.IsLinkLocalMulticast() || r.IsLinkLocalUnicast() {
						continue
					}
					if p := r.To4(); p != nil {
						l = append(l, Address(p.String()))
					} else if IPv6 {
						l = append(l, Address(r.String()))
					}
				}
			}
		}
	}
	if len(l) > 256 {
		return l[:254]
	}
	return l
}

func (d *Device) MarshalStream(w dio.Writer) error {
	if _, err := w.WriteString(d.ID); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(d.OS)); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(d.Arch)); err != nil {
		return err
	}
	if err := w.WriteBool(d.Admin); err != nil {
		return err
	}
	if _, err := w.WriteString(d.Family); err != nil {
		return err
	}
	if _, err := w.WriteString(d.Version); err != nil {
		return err
	}
	if _, err := w.WriteString(d.Hostname); err != nil {
		return err
	}
	if err := w.WriteUint8(uint8(len(d.IPs))); err != nil {
		return err
	}
	for _, a := range d.IPs {
		if _, err := w.WriteString(string(a)); err != nil {
			return err
		}
	}
	return nil
}
func (d *Device) UnmarshalStream(r dio.Reader) error {
	return nil
}
