// Copyright (C) 2020 - 2022 iDigitalFlame
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
	"net/netip"

	"github.com/iDigitalFlame/xmt/data"
)

// Address represents an encoded IPv4 or IPv6 address.
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

// String returns the string form of the IP address.
func (a Address) String() string {
	if a.IsUnspecified() {
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
// address "0.0.0.0" or the IPv6 address "::".
func (a Address) IsUnspecified() bool {
	return a.hi == 0 && a.low == 0
}

// ToAddr will return this Address as a netip.Addr struct. This will choose the
// type based on the underlying address size.
func (a *Address) ToAddr() netip.Addr {
	if a.Is4() {
		return netip.AddrFrom4([4]byte{byte(a.low >> 24), byte(a.low >> 16), byte(a.low >> 8), byte(a.low)})
	}
	return netip.AddrFrom16([16]byte{
		byte(a.hi >> 56), byte(a.hi >> 48), byte(a.hi >> 40), byte(a.hi >> 32),
		byte(a.hi >> 24), byte(a.hi >> 16), byte(a.hi >> 8), byte(a.hi),
		byte(a.low >> 56), byte(a.low >> 48), byte(a.low >> 40), byte(a.low >> 32),
		byte(a.low >> 24), byte(a.low >> 16), byte(a.low >> 8), byte(a.low),
	})
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
	return !a.IsUnspecified() && !a.IsBroadcast() && !a.IsLoopback() && !a.IsMulticast() && !a.IsLinkLocalUnicast()
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
