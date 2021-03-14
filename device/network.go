package device

import (
	"net"
	"strings"
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

const maxNetworks = 255

type device struct {
	Name    string    `json:"name"`
	Address []address `json:"address"`
	Mac     hardware  `json:"mac"`
}
type hardware uint64

// Network is a basic listing of network interfaces.  Used to store and refresh interface lists.
type Network []device

// The address struct was built on the great work from the great inet.af/netaddr package
// thanks and great work y'all! Godoc: https://pkg.go.dev/inet.af/netaddr
// https://tailscale.com/blog/netaddr-new-ip-type-for-go/
type address struct {
	hi, low uint64
}

// Len returns the number of detected interfaces detected.
func (n Network) Len() int {
	return len(n)
}
func (a address) Len() int {
	if a.Is4() {
		return 32
	}
	return 128
}
func (a address) Is4() bool {
	return a.hi == 0 && a.low>>32 == 0xFFFF
}
func (a address) Is6() bool {
	return a.low > 0 && a.low>>32 != 0xFFFF
}
func (a address) IP() net.IP {
	if a.Is4() {
		return net.IP{byte(a.low >> 24), byte(a.low >> 16), byte(a.low >> 8), byte(a.low)}
	}
	return net.IP{
		byte(a.hi >> 56), byte(a.hi >> 48), byte(a.hi >> 40), byte(a.hi >> 32),
		byte(a.hi >> 24), byte(a.hi >> 16), byte(a.hi >> 8), byte(a.hi),
		byte(a.low >> 56), byte(a.low >> 48), byte(a.low >> 40), byte(a.low >> 32),
		byte(a.low >> 24), byte(a.low >> 16), byte(a.low >> 8), byte(a.low),
	}
}
func (a address) IsZero() bool {
	return a.hi == 0 && a.low == 0
}
func (d device) String() string {
	b := builders.Get().(*strings.Builder)
	b.WriteString(d.Name + "(" + d.Mac.String() + "): [")
	for i := range d.Address {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(d.Address[i].String())
	}
	b.WriteByte(']')
	r := b.String()
	b.Reset()
	builders.Put(b)
	return r
}
func (a *address) set(i net.IP) {
	if len(i) == 4 {
		_ = i[3]
		a.hi, a.low = 0, 0xFFFF00000000|uint64(i[0])<<24|uint64(i[1])<<16|uint64(i[2])<<8|uint64(i[3])
		return
	}
	_ = i[15]
	a.hi = uint64(i[7]) | uint64(i[6])<<8 | uint64(i[5])<<16 | uint64(i[4])<<24 |
		uint64(i[3])<<32 | uint64(i[2])<<40 | uint64(i[1])<<48 | uint64(i[0])<<56
	a.low = uint64(i[15]) | uint64(i[14])<<8 | uint64(i[13])<<16 | uint64(i[12])<<24 |
		uint64(i[11])<<32 | uint64(i[10])<<40 | uint64(i[9])<<48 | uint64(i[8])<<56
}
func (a address) String() string {
	b := builders.Get().(*strings.Builder)
	if a.Is4() {
		b.Grow(15)
		write(b, uint8(a.low>>24))
		b.WriteByte('.')
		write(b, uint8(a.low>>16))
		b.WriteByte('.')
		write(b, uint8(a.low>>8))
		b.WriteByte('.')
		write(b, uint8(a.low))
	} else {
		var s, e, i uint8 = 255, 255, 0
		for b.Grow(39); i < 8; i++ {
			j := i
			for j < 8 && a.grab(j) == 0 {
				j++
			}
			if l := j - i; l >= 2 && l > e-s {
				s, e = i, j
			}
		}
		for i = 0; i < 8; i++ {
			if i == s {
				b.WriteString("::")
				if i = e; i >= 8 {
					break
				}
			} else if i > 0 {
				b.WriteByte(':')
			}
			writeHex(b, a.grab(i))
		}
	}
	r := b.String()
	b.Reset()
	builders.Put(b)
	return r
}
func (h hardware) String() string {
	b := builders.Get().(*strings.Builder)
	b.Grow(17)
	for n, i := uint8(0), uint8(0); i < 6; i++ {
		if n = uint8(h >> ((5 - i) * 8)); i > 0 {
			b.WriteByte(':')
		}
		b.WriteByte(table[n>>4])
		b.WriteByte(table[n&0x0F])
	}
	r := b.String()
	b.Reset()
	builders.Put(b)
	return r
}

// Refresh collects the interfaces connected to this system and fills this Network object with the information.
// If previous Network information is contained in this Network object, it is cleared before filling.
func (n *Network) Refresh() error {
	if len(*n) > 0 {
		*n = (*n)[0:0]
	}
	l, err := net.Interfaces()
	if err != nil {
		return xerr.Wrap("cannot get interfaces", err)
	}
	for i := range l {
		if l[i].Flags&net.FlagUp == 0 || l[i].Flags&net.FlagLoopback != 0 || l[i].Flags&net.FlagPointToPoint != 0 {
			continue
		}
		a, err := l[i].Addrs()
		if err != nil || len(a) == 0 {
			continue
		}
		c := device{Name: l[i].Name, Address: make([]address, 0, len(a)), Mac: mac(l[i].HardwareAddr)}
		for o := range a {
			if o > maxNetworks {
				break
			}
			var t address
			switch z := a[o].(type) {
			case *net.IPNet:
				t.set(z.IP)
			case *net.IPAddr:
				t.set(z.IP)
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
func (a address) IsLoopback() bool {
	return (a.Is4() && uint8(a.low>>24) == 0x7F) || (a.Is6() && a.hi == 0 && a.low == 1)
}
func (a address) IsMulticast() bool {
	return (a.Is4() && uint8(a.low>>24) == 0xFE) || (a.Is6() && uint8(a.hi>>56) == 0xFF)
}
func (a address) IsBroadcast() bool {
	return a.Is4() && a.low == 0xFFFFFFFFFFFF
}
func mac(h net.HardwareAddr) hardware {
	_ = h[5]
	return hardware(
		uint64(h[0])<<40 | uint64(h[1])<<32 | uint64(h[2])<<24 |
			uint64(h[3])<<16 | uint64(h[4])<<8 | uint64(h[5]),
	)
}
func (a address) grab(i uint8) uint16 {
	r := a.hi
	if (i/4)%2 == 1 {
		r = a.low
	}
	return uint16(r >> ((3 - i%4) * 16))
}
func (a address) IsUnspecified() bool {
	return a.hi == 0 && a.low == 0
}
func (a address) IsGlobalUnicast() bool {
	return !a.IsBroadcast() && !a.IsUnspecified() && !a.IsLoopback() && !a.IsMulticast() && !a.IsLinkLocalUnicast()
}
func write(b *strings.Builder, v uint8) {
	if v >= 100 {
		b.WriteByte(table[v/100])
	}
	if v >= 10 {
		b.WriteByte(table[v/10%10])
	}
	b.WriteByte(table[v%10])
}
func (a address) IsLinkLocalUnicast() bool {
	return (a.Is4() && uint32(a.low>>16) == 0xFFFFA9FE) || (a.Is6() && uint16(a.hi>>48) == 0xFE80)
}
func writeHex(b *strings.Builder, v uint16) {
	if v >= 0x1000 {
		b.WriteByte(table[v>>12])
	}
	if v >= 0x100 {
		b.WriteByte(table[v>>8&0xF])
	}
	if v >= 0x10 {
		b.WriteByte(table[v>>4&0xF])
	}
	b.WriteByte(table[v&0xF])
}
func (a address) IsLinkLocalMulticast() bool {
	return (a.Is4() && a.low>>8 == 0xFFFFE00000) || (a.Is6() && uint16(a.hi>>48) == 0xFF02)
}
func (a address) IsInterfaceLocalMulticast() bool {
	return a.Is6() && uint8(a.hi>>56) == 0xFF && uint8(a.hi>>48) == 1
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
func (a address) MarshalStream(w data.Writer) error {
	if err := w.WriteUint64(a.hi); err != nil {
		return err
	}
	return w.WriteUint64(a.low)
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
	d.Address = make([]address, l)
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
	for ; l > 0; l-- {
		var d device
		if err := d.UnmarshalStream(r); err != nil {
			return err
		}
		*n = append(*n, d)
	}
	return nil
}
func (a *address) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint64(&a.hi); err != nil {
		return err
	}
	return r.ReadUint64(&a.low)
}
func (h *hardware) UnmarshalStream(r data.Reader) error {
	return r.ReadUint64((*uint64)(unsafe.Pointer(h)))
}
