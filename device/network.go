package device

import (
	"fmt"
	"net"

	"github.com/iDigitalFlame/xmt/data"
)

const maxNetworks = 255

type device struct {
	Name     string           `json:"name"`
	Address  []net.IP         `json:"address"`
	Hardware net.HardwareAddr `json:"mac"`
}

// Network is a basic listing of network interfaces.  Used to store and refresh interface lists.
type Network []*device

// Len returns the number of detected interfaces detected.
func (n Network) Len() int {
	return len(n)
}
func (d device) String() string {
	return fmt.Sprintf(
		"%s (%s): %s", d.Name, d.Hardware.String(), d.Address,
	)
}

// Refresh collects the interfaces connected to this system and fills this Network object with the information.
// If previous Network information is contained in this Network object, it is cleared before filling.
func (n Network) Refresh() error {
	if len(n) > 0 {
		for i := range n {
			n[i] = nil
		}
		n = n[0:0]
	}
	l, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("error retriving interfaces: %w", err)
	}
	for i := range l {
		if l[i].Flags&net.FlagUp == 0 || l[i].Flags&net.FlagLoopback != 0 || l[i].Flags&net.FlagPointToPoint != 0 {
			continue
		}
		a, err := l[i].Addrs()
		if err != nil || len(a) == 0 {
			continue
		}
		c := &device{
			Name:     l[i].Name,
			Address:  make([]net.IP, 0),
			Hardware: l[i].HardwareAddr,
		}
		for o := range a {
			if o > maxNetworks {
				break
			}
			var t net.IP
			switch a[o].(type) {
			case *net.IPNet:
				t = a[o].(*net.IPNet).IP
			case *net.IPAddr:
				t = a[o].(*net.IPAddr).IP
			default:
				continue
			}
			if t.IsLoopback() || t.IsUnspecified() || t.IsMulticast() || t.IsInterfaceLocalMulticast() || t.IsLinkLocalMulticast() || t.IsLinkLocalUnicast() {
				continue
			}
			if p := t.To4(); p != nil {
				c.Address = append(c.Address, p)
			} else if IPv6 {
				c.Address = append(c.Address, t)
			}
		}
		if len(c.Address) > 0 {
			n = append(n, c)
		}
	}
	return nil
}
func (d device) MarshalStream(w data.Writer) error {
	if err := w.WriteString(d.Name); err != nil {
		return err
	}
	if err := w.WriteBytes(d.Hardware); err != nil {
		return err
	}
	l := uint8(len(d.Address))
	if err := w.WriteUint8(l); err != nil {
		return err
	}
	for x := uint8(0); x < l; x++ {
		if err := w.WriteBytes(d.Address[x]); err != nil {
			return err
		}
	}
	return nil
}

// MarshalStream writes the data of this Network to the supplied Writer.
func (n Network) MarshalStream(w data.Writer) error {
	l := uint8(len(n))
	if err := w.WriteUint8(l); err != nil {
		return err
	}
	for x := uint8(0); x < l; x++ {
		if err := n[x].MarshalStream(w); err != nil {
			return err
		}
	}
	return nil
}
func (d *device) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&d.Name); err != nil {
		return err
	}
	m, err := r.Bytes()
	if err != nil {
		return err
	}
	d.Hardware = net.HardwareAddr(m)
	l, err := r.Uint8()
	if err != nil {
		return err
	}
	d.Address = make([]net.IP, l)
	for x := uint8(0); x < l; x++ {
		a, err := r.Bytes()
		if err != nil {
			return err
		}
		d.Address[x] = net.IP(a)
	}
	return nil
}

// UnmarshalStream reads the data of this Network from the supplied Reader.
func (n *Network) UnmarshalStream(r data.Reader) error {
	l, err := r.Uint8()
	if err != nil {
		return err
	}
	for ; l > 0; l-- {
		d := new(device)
		if err := d.UnmarshalStream(r); err != nil {
			return err
		}
		*n = append(*n, d)
	}
	return nil
}
