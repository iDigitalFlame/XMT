package device

import (
	"net"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Address represents an encoded IPv4 or IPv6 address.
// NOTE(dij): Might get replaced in Go1.18 with netip.Address
//
// The address struct was built on the great work from the great inet.af/netaddr
// package thanks and great work y'all!
//
// GoDoc: https://pkg.go.dev/inet.af/netaddr
//
// https://tailscale.com/blog/netaddr-new-ip-type-for-go/
type Address struct {
	hi, low uint64
}

// Len returns the size of this IP address. It returns '32' for IPv4 and '128'
// for IPv6.
func (a Address) Len() int {
	if a.Is4() {
		return 32
	}
	return 128
}

// Is4 returns true if this struct represents an IPv4 based address or an IPv4
// address wrapped in an IPv6 address.
func (a Address) Is4() bool {
	return a.hi == 0 && a.low>>32 == 0xFFFF
}

// Is6 returns true if this struct represents an IPv6 based address.
func (a Address) Is6() bool {
	return a.low > 0 && a.low>>32 != 0xFFFF
}

// IP returns a 'net.IP' copy of this address.
//
// This may be zero or empty depending on the type of address value this struct
// contains.
func (a Address) IP() net.IP {
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

// FromIP  will create a new Address struct and set it's contents based on the
// value of the supplied 'net.IP'.
func FromIP(i net.IP) *Address {
	var a Address
	a.Set(i)
	return &a
}

// IsZero returns true if this struct represents an empty or unset address.
func (a Address) IsZero() bool {
	return a.hi == 0 && a.low == 0
}

// Set will set the internal values of this address to the specified 'net.IP'
// address.
func (a *Address) Set(i net.IP) {
	if len(i) == 0 {
		return
	}
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

// ParseIP parses s as an IP address, returning the result. The string s can be
// in IPv4 dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"), or IPv4-mapped
// IPv6 ("::ffff:192.0.2.1") form.
//
// If s is not a valid textual representation of an IP address, ParseIP returns
// nil.
func ParseIP(s string) *Address {
	i := net.ParseIP(s)
	if len(i) == 0 {
		return nil
	}
	var a Address
	a.Set(i)
	return &a
}

// String returns the string form of the IP address.
func (a Address) String() string {
	if a.IsZero() {
		if a.Is4() {
			return emptyIP
		}
		return "::"
	}
	var n int
	if a.Is4() {
		var b [15]byte
		n = write(&b, uint8(a.low>>24), n)
		b[n] = '.'
		n++
		n = write(&b, uint8(a.low>>16), n)
		b[n] = '.'
		n++
		n = write(&b, uint8(a.low>>8), n)
		b[n] = '.'
		n++
		n = write(&b, uint8(a.low), n)
		return string(b[:n])
	}
	var (
		b       [39]byte
		s, e, i uint8 = 255, 255, 0
	)
	for ; i < 8; i++ {
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
			b[n] = ':'
			b[n+1] = ':'
			n += 2
			if i = e; i >= 8 {
				break
			}
		} else if i > 0 {
			b[n] = ':'
			n++
		}
		n = writeHex(&b, a.grab(i), n)
	}
	return string(b[:n])
}

// IsPrivate reports whether ip is a private address, according to RFC 1918
// (IPv4 addresses) and RFC 4193 (IPv6 addresses).
func (a Address) IsPrivate() bool {
	if a.Is4() {
		f, s := byte(a.low>>24), byte(a.low>>16)
		return f == 10 || (f == 172 && s&0xf0 == 16) || (f == 192 && s == 168)
	}
	return byte(a.hi>>56)&0xFE == 0xFC
}

// IsLoopback reports whether this is a loopback address.
func (a Address) IsLoopback() bool {
	return (a.Is4() && uint8(a.low>>24) == 0x7F) || (a.Is6() && a.hi == 0 && a.low == 1)
}

// IsMulticast reports whether this is a multicast address.
func (a Address) IsMulticast() bool {
	return (a.Is4() && uint8(a.low>>24) == 0xFE) || (a.Is6() && uint8(a.hi>>56) == 0xFF)
}

// IsBroadcast reports whether this is a broadcast address.
func (a Address) IsBroadcast() bool {
	return a.Is4() && a.low == 0xFFFFFFFFFFFF
}
func (a Address) grab(i uint8) uint16 {
	r := a.hi
	if (i/4)%2 == 1 {
		r = a.low
	}
	return uint16(r >> ((3 - i%4) * 16))
}

// IsUnspecified reports whether ip is an unspecified address, either the IPv4
// address "0.0.0.0" or the IPv6 address "::". Same as 'IsZero'.
func (a Address) IsUnspecified() bool {
	return a.hi == 0 && a.low == 0
}

// IsGlobalUnicast reports whether this is a global unicast address.
//
// The identification of global unicast addresses uses address type identification
// as defined in RFC 1122, RFC 4632 and RFC 4291 with the exception of IPv4
// directed broadcast addresses.
//
// It returns true even if this is in IPv4 private address space or local IPv6
// unicast address space.
func (a Address) IsGlobalUnicast() bool {
	return !a.IsBroadcast() && !a.IsUnspecified() && !a.IsLoopback() && !a.IsMulticast() && !a.IsLinkLocalUnicast()
}

// IsLinkLocalUnicast reports whether this is a link-local unicast address.
func (a Address) IsLinkLocalUnicast() bool {
	return (a.Is4() && uint32(a.low>>16) == 0xFFFFA9FE) || (a.Is6() && uint16(a.hi>>48) == 0xFE80)
}
func write(b *[15]byte, v uint8, n int) int {
	if v >= 100 {
		b[n] = table[v/100]
		n++
	}
	if v >= 10 {
		b[n] = table[v/10%10]
		n++
	}
	b[n] = table[v%10]
	return n + 1
}

// IsLinkLocalMulticast reports whether this is a link-local multicast address.
func (a Address) IsLinkLocalMulticast() bool {
	return (a.Is4() && a.low>>8 == 0xFFFFE00000) || (a.Is6() && uint16(a.hi>>48) == 0xFF02)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (a Address) MarshalText() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + a.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *Address) UnmarshalJSON(b []byte) error {
	if len(b) < 1 || b[len(b)-1] != '"' || b[0] != '"' {
		if xerr.Concat {
			return xerr.Sub(`invalid address value "`+string(b)+`"`, 0x90)
		}
		return xerr.Sub("invalid address value", 0xD)
	}
	if i := net.ParseIP(string(b[1 : len(b)-2])); i != nil {
		a.Set(i)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (a *Address) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	if i := net.ParseIP(string(b)); i != nil {
		a.Set(i)
	}
	return nil
}
func writeHex(b *[39]byte, v uint16, n int) int {
	if v >= 0x1000 {
		b[n] = table[v>>12]
		n++
	}
	if v >= 0x100 {
		b[n] = table[v>>8&0xF]
		n++
	}
	if v >= 0x10 {
		b[n] = table[v>>4&0xF]
		n++
	}
	b[n] = table[v&0xF]
	return n + 1
}

// IsInterfaceLocalMulticast reports whether this is an interface-local multicast
// address.
func (a Address) IsInterfaceLocalMulticast() bool {
	return a.Is6() && uint8(a.hi>>56) == 0xFF && uint8(a.hi>>48) == 1
}

// MarshalStream writes the data of this Address to the supplied Writer.
func (a Address) MarshalStream(w data.Writer) error {
	if err := w.WriteUint64(a.hi); err != nil {
		return err
	}
	return w.WriteUint64(a.low)
}

// UnmarshalStream reads the data of this Address from the supplied Reader.
func (a *Address) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint64(&a.hi); err != nil {
		return err
	}
	return r.ReadUint64(&a.low)
}
