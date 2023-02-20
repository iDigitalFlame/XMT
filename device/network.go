// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package device

import (
	"net"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util"
)

const maxNetworks = 255

type device struct {
	Name    string
	Address []Address
	Mac     hardware
}
type hardware uint64

// Network is a basic listing of network interfaces. Used to store and refresh
// interface lists.
type Network []device

// Len returns the number of detected interfaces detected.
func (n Network) Len() int {
	return len(n)
}
func (h hardware) String() string {
	var (
		b [17]byte
		v int
	)
	for n, i := uint8(0), uint8(0); i < 6; i++ {
		if n = uint8(h >> ((5 - i) * 8)); i > 0 {
			b[v] = ':'
			v++
		}
		b[v] = util.HexTable[n>>4]
		b[v+1] = util.HexTable[n&0x0F]
		v += 2
	}
	return string(b[:v])
}

// Refresh collects the interfaces connected to this system and fills this
// Network object with the information.
//
// If previous Network information is contained in this Network object, it is
// cleared before filling.
func (n *Network) Refresh() error {
	if len(*n) > 0 {
		*n = (*n)[0:0]
	}
	l, err := net.Interfaces()
	if err != nil {
		return err
	}
	for i := range l {
		if l[i].Flags&net.FlagUp == 0 || l[i].Flags&net.FlagLoopback != 0 || l[i].Flags&net.FlagPointToPoint != 0 {
			continue
		}
		a, err := l[i].Addrs()
		if err != nil || len(a) == 0 {
			continue
		}
		c := device{Name: l[i].Name, Address: make([]Address, 0, len(a)), Mac: mac(l[i].HardwareAddr)}
		for o := range a {
			if o > maxNetworks {
				break
			}
			var t Address
			switch z := a[o].(type) {
			case *net.IPNet:
				t.Set(z.IP)
			case *net.IPAddr:
				t.Set(z.IP)
			default:
				continue
			}
			if !t.IsGlobalUnicast() {
				continue
			}
			c.Address = append(c.Address, t)
		}
		if len(c.Address) > 0 {
			*n = append(*n, c)
		}
	}
	return nil
}
func mac(h net.HardwareAddr) hardware {
	_ = h[5]
	return hardware(
		uint64(h[0])<<40 | uint64(h[1])<<32 | uint64(h[2])<<24 |
			uint64(h[3])<<16 | uint64(h[4])<<8 | uint64(h[5]),
	)
}
func (d device) MarshalStream(w data.Writer) error {
	if err := w.WriteString(d.Name); err != nil {
		return err
	}
	if err := d.Mac.MarshalStream(w); err != nil {
		return err
	}
	l := uint8(len(d.Address))
	if err := w.WriteUint8(l); err != nil {
		return err
	}
	for x := uint8(0); x < l; x++ {
		if err := d.Address[x].MarshalStream(w); err != nil {
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
func (h hardware) MarshalStream(w data.Writer) error {
	return w.WriteUint64(uint64(h))
}
func (d *device) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&d.Name); err != nil {
		return err
	}
	if err := d.Mac.UnmarshalStream(r); err != nil {
		return err
	}
	l, err := r.Uint8()
	if err != nil {
		return err
	}
	d.Address = make([]Address, l)
	for x := uint8(0); x < l; x++ {
		if err := d.Address[x].UnmarshalStream(r); err != nil {
			return err
		}
	}
	return nil
}

// UnmarshalStream reads the data of this Network from the supplied Reader.
func (n *Network) UnmarshalStream(r data.Reader) error {
	l, err := r.Uint8()
	if err != nil {
		return err
	}
	*n = make(Network, l)
	for x := uint8(0); x < l; x++ {
		if err := (*n)[x].UnmarshalStream(r); err != nil {
			return err
		}
	}
	return nil
}
func (h *hardware) UnmarshalStream(r data.Reader) error {
	return r.ReadUint64((*uint64)(h))
}
