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

package crypto

import (
	"crypto/rand"
	"io"
	"sync"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

const size = 128

var (
	chains = sync.Pool{
		New: func() any {
			var b [size + 1]byte
			return &b
		},
	}
	tables = sync.Pool{
		New: func() any {
			var b [size + 1][256]byte
			return &b
		},
	}
)

// CBK is the representation of the CBK Cipher.
// CBK is a block based cipher that allows for a variable size index in encoding.
type CBK struct {
	// Random Source to use for data generation from keys.
	// This source MUST be repeatable.
	Source     source
	buf        []byte
	pos, total int

	A, B  byte
	C, D  byte
	index uint8
}
type source interface {
	Seed(int64)
	Int63() int64
}

// NewCBK returns a new CBK Cipher with the D value specified. The other A, B and
// C values are randomly generated at runtime.
func NewCBK(d int) *CBK {
	c, _ := NewCBKEx(d, size, nil)
	return c
}

// Reset resets the encryption keys and sets them to new random bytes.
func (e *CBK) Reset() error {
	if _, err := rand.Read(e.buf[0:3]); err != nil {
		return err
	}
	_ = e.buf[2]
	e.A, e.B, e.C = e.buf[0], e.buf[1], e.buf[2]
	e.pos, e.index = 0, 0
	return nil
}

// BlockSize returns the cipher's block BlockSize.
func (e *CBK) BlockSize() int {
	return len(e.buf) - 1
}

// Shuffle will switch around the bytes in the array based on the Cipher bytes.
func (e *CBK) Shuffle(b []byte) {
	if len(b) > 1 {
		b[0] += e.A
	}
	for i := byte(0); i < byte(len(b)); i++ {
		switch {
		case i%e.A == 0:
			b[i] += e.A - i
		case e.C%i == 0:
			b[i] += e.B - e.D
		case i == e.D:
			b[i] -= e.A + i
		default:
			if i%2 == 0 {
				b[i] += e.B / 3
			} else {
				b[i] += e.C / 5
			}
		}
	}
}

// Deshuffle will reverse the switch around the bytes in the array based on the
// Cipher bytes.
func (e *CBK) Deshuffle(b []byte) {
	if len(b) > 1 {
		b[0] -= e.A
	}
	for i := byte(0); i < byte(len(b)); i++ {
		switch {
		case i%e.A == 0:
			b[i] -= e.A - i
		case e.C%i == 0:
			b[i] -= e.B - e.D
		case i == e.D:
			b[i] += e.A + i
		default:
			if i%2 == 0 {
				b[i] -= e.B / 3
			} else {
				b[i] -= e.C / 5
			}
		}
	}
}
func (e *CBK) adjust(i uint16) uint16 {
	if e.Source != nil {
		return uint16(e.Source.Int63() * int64(i+1))
	}
	if n := ((uint16(e.A) ^ uint16(e.B)) - uint16(e.C)) * (i + 1); n > 1 {
		return n
	}
	return 1
}

// Encrypt encrypts the first block in src into dst. Dst and src must overlap entirely
// or not at all.
func (e *CBK) Encrypt(dst, src []byte) {
	copy(dst, src)
	e.Shuffle(dst)
	e.scramble(dst, true)
}

// Decrypt decrypts the first block in src into dst. Dst and src must overlap entirely
// or not at all.
func (e *CBK) Decrypt(dst, src []byte) {
	copy(dst, src)
	e.scramble(dst, false)
	e.Deshuffle(dst)
}

// Flush pushes the remaining bytes stored into the buffer into the supplies Writer.
func (e *CBK) Flush(w io.Writer) error {
	_, err := e.flushOutput(w)
	return err
}
func (e *CBK) scramble(b []byte, d bool) {
	var (
		o    = chains.Get().(*[size + 1]byte)
		x    = e.adjust(uint16(e.A*e.B) + uint16(e.D))
		y    = e.adjust(uint16((e.C-e.D)*e.A) + x + e.adjust(uint16(e.index)))
		z    = e.adjust(uint16(byte(x*y) + e.B - e.D*e.index))
		i    int8
		g, h byte
	)
	if d {
		i = 5
	}
	for (i < 6 && !d) || (i >= 0 && d) {
		g = (byte(z*y) + e.blockIndex(true, uint16(e.D*e.A)+uint16(i)+x, uint16(e.D)+uint16(e.index))) % 8
		h = (byte(y) - e.blockIndex(false, y+uint16(e.D)+uint16(e.index*uint8(i+1)), uint16(e.D)+x+uint16(byte(uint16(i)*z)*e.A))) % 8
		if g != h {
			if !d {
				b[h], b[g] = (b[g]&0xF)<<4|(b[h]&0xF), (b[g]>>4)<<4|((b[h]>>4)&0xF)
				b[h+1], b[g+1] = (b[g+1]&0xF)<<4|(b[h+1]&0xF), (b[g+1]>>4)<<4|((b[h+1]>>4)&0xF)
			}
			copy((*o)[0:2], b[g*2:(g*2)+2])
			copy(b[g*2:], b[h*2:(h*2)+2])
			copy(b[h*2:], (*o)[0:2])
			if d {
				b[h], b[g] = (b[g]&0xF)<<4|(b[h]&0xF), (b[g]>>4)<<4|((b[h]>>4)&0xF)
				b[h+1], b[g+1] = (b[g+1]&0xF)<<4|(b[h+1]&0xF), (b[g+1]>>4)<<4|((b[h+1]>>4)&0xF)
			}
		}
		if d {
			i--
		} else {
			i++
		}
	}
	clear(o, nil)
}
func (e *CBK) cipherTable(b *[size + 1]byte) {
	(*b)[0] = byte(uint16(e.index+1)*uint16(e.D+1) + e.adjust(uint16(e.D)))
	for i := byte(1); i < byte(len(*b))-1; i++ {
		switch {
		case i <= 6:
			if i%2 == 0 {
				(*b)[i] = byte(uint16(e.index) - uint16(e.A) + uint16(e.B-(i-e.C)) + uint16(i) - e.adjust(uint16(e.A)))
			} else {
				(*b)[i] = byte(uint16(e.index) - uint16(e.A) + uint16(e.B-(i-3)) + uint16(i) - e.adjust(uint16(e.A)))
			}
		case i > 6 && i <= 11:
			(*b)[i] = byte(uint16(e.C) - uint16(e.B) + uint16((e.index+1)*i) + e.adjust(uint16(e.C)))
		case i > 11:
			(*b)[i] = byte(e.adjust(uint16(e.B+e.C)) + uint16(e.D) - uint16(len(*b)-1) - uint16(e.D) + uint16(e.A-e.C))
		}
	}
	(*b)[len(*b)-1] = byte(e.adjust(uint16(e.B+e.C)) + uint16(e.index) - uint16(len(*b)-1) - uint16(e.D) + uint16(e.A-e.C))
}
func (e *CBK) readInput(r io.Reader) (int, error) {
	n, err := io.ReadFull(r, e.buf)
	if n <= 0 {
		if e.total = 0; err == nil {
			return 0, io.EOF
		}
		return 0, err
	}
	if n != len(e.buf) {
		return 0, io.ErrUnexpectedEOF
	}
	if e.index++; e.index > 30 {
		e.index = 0
	}
	var (
		t = chains.Get().(*[size + 1]byte)
		c = tables.Get().(*[size + 1][256]byte)
	)
	e.cipherTable(t)
	e.Deshuffle(e.buf)
	e.scramble(e.buf, true)
	for x := range *c {
		for z := range (*c)[x] {
			(*c)[x][(*t)[x]&0xFF] = byte(z)
			(*t)[x]++
		}
	}
	for i := range e.buf {
		e.buf[i] = (*c)[i&0xF][e.buf[i]&0xFF]
	}
	e.total, e.pos = int(e.buf[len(e.buf)-1]), 0
	if clear(t, c); e.total == 0 {
		return 0, io.EOF
	}
	if e.total > len(e.buf)-1 {
		return n, io.ErrShortBuffer
	}
	return n, err
}
func (e *CBK) blockIndex(a bool, t, i uint16) byte {
	switch v := t % 8; {
	case v == 0 && a:
		return byte((((t+1)*(1+i+uint16(e.A)*t) + t + 5) / 3) + 4 + (5 * t) + (i / 5))
	case v == 1 && a:
		return byte((t / 5) + i + ((i + 1) * 7) + ((1 + t) * 3) + (i / 2) + t)
	case v == 2 && a:
		return byte((((3+t+uint16(e.B+e.C))/4+1)+i)/2 + (3 * t) + (t / 5) + i + 3)
	case v == 3 && a:
		return byte(((t / 2) * 3) + 7 + ((t + i) * 3) - 2 + ((t * (i + 5 + uint16(e.D))) * 3))
	case v == 4 && a:
		return byte((((i*6)+2)/5)*3 + ((4 * i) / 5) + 3 + (t / 4))
	case v == 5 && a:
		return byte((((t*3)/5)+(5+i))*3 + (t * (2 - uint16(e.A*e.D))) + (i / (t + 1)) + (6 + t))
	case v == 6 && a:
		return byte((((((i + 5) / 3) * 7) + 3 + uint16(e.B)) / (t + 1)) + 3 + (t/(i+1))*3)
	case v == 7 && a:
		return byte(((((t / (i + 1) * 2) + 5) / 4) + 10) + (3 * t) + ((i / 2) + (t * 3)) + 4)
	case v == 0 && !a:
		return byte((((3/(2+i) + 3) / (t + 1)) * 9) + 6 - uint16(e.A*e.C) + i)
	case v == 1 && !a:
		return byte(((((4*i)/3 + (t * 2)) / 3) + 8) / 3)
	case v == 2 && !a:
		return byte((((9 + i + uint16(e.A*e.D)) / 4) + (t / 2) + (2*i + 1 + uint16(e.D))) / (((i + 3) / (5 + t)) + 6))
	case v == 3 && !a:
		return byte(((((4+(t-5)/2)/6)+3)*2)*((5+i)/3) + 4)
	case v == 4 && !a:
		return byte((((((t/3)/(3+i) + uint16(e.C)) / 9) * 2) + 8) + (5+i)/(3+t))
	case v == 5 && !a:
		return byte(((i * 4) + (t / 3) - uint16(e.A*byte(1+t)) + (6 / (1 + i))) + (6 / (3 + t)) + (i * 3))
	case v == 6 && !a:
		return byte((((((t*9)/6)+(i*3)/9)*5 + i) - uint16(e.D*byte(i))) + (t+2)/4)
	case v == 7 && !a:
		return byte((((((i/3)*7)+3-uint16(e.B))*5 + t) * (t + 3) / 7) + uint16(e.D*e.B))
	}
	return 0
}
func (e *CBK) flushOutput(w io.Writer) (int, error) {
	if e.pos == 0 {
		return 0, nil
	}
	if e.index++; e.index > 30 {
		e.index = 0
	}
	e.buf[e.total] = byte(e.pos)
	var (
		t = chains.Get().(*[size + 1]byte)
		c = tables.Get().(*[size + 1][256]byte)
	)
	e.cipherTable(t)
	for x := range *c {
		for z := range (*c)[x] {
			(*c)[x][z] = (*t)[x]
			(*t)[x]++
		}
	}
	for i := range e.buf {
		e.buf[i] = (*c)[i&0xF][e.buf[i]&0xFF]
	}
	e.scramble(e.buf, false)
	e.Shuffle(e.buf)
	e.pos = 0
	n, err := w.Write(e.buf)
	clear(t, c)
	return n, err
}

// NewCBKSource returns a new CBK Cipher with the A, B, C, D, BlockSize values
// specified.
func NewCBKSource(a, b, c, d, sz byte) (*CBK, error) {
	switch sz {
	case 0:
		sz = size
	case 16, 32, 64, 128:
	default:
		return nil, xerr.Sub("block size must be a power of two between 16 and 128", 0x28)
	}
	return &CBK{A: a, B: b, C: c, D: d, buf: make([]byte, sz+1), total: -1}, nil
}
func clear(b *[size + 1]byte, z *[size + 1][256]byte) {
	for i := range *b {
		(*b)[i] = 0
	}
	if chains.Put(b); z != nil {
		tables.Put(z)
	}
}

// NewCBKEx returns a new CBK Cipher with the D value, BlockSize and Entropy source
// specified. The other A, B and C values are randomly generated at runtime.
func NewCBKEx(d int, sz int, src source) (*CBK, error) {
	switch sz {
	case 0:
		sz = size
	case 16, 32, 64, 128:
	default:
		return nil, xerr.Sub("block size must be a power of two between 16 and 128", 0x28)
	}
	c := &CBK{D: byte(d), buf: make([]byte, sz+1), total: -1, Source: src}
	c.Reset()
	return c, nil
}

// Read reads the contents of the Reader to the byte array after decrypting with
// this Cipher.
func (e *CBK) Read(r io.Reader, b []byte) (int, error) {
	if e.buf == nil {
		e.buf = make([]byte, size+1)
	}
	if e.total-e.pos > len(b) {
		if e.pos+len(b) > len(e.buf) {
			return 0, io.ErrShortBuffer
		}
		u := copy(b, e.buf[e.pos:e.pos+len(b)])
		e.pos += len(b)
		return u, nil
	}
	if e.pos >= e.total {
		if o, err := e.readInput(r); err != nil && (err != io.EOF || o == 0) {
			return o, err
		}
	}
	var n int
	for i := 0; n < len(b) && e.pos < e.total && e.total < len(e.buf); n += i {
		if e.total <= 0 {
			return n, io.EOF
		}
		i = copy(b[n:], e.buf[e.pos:e.total])
		if e.pos += i; e.pos >= e.total && e.total >= len(e.buf)-1 {
			if _, err := e.readInput(r); err != nil && err != io.EOF {
				return n, err
			}
		}
	}
	if e.total > len(e.buf) {
		return n, io.EOF
	}
	return n, nil
}

// Write writes the contents of the byte array to the Writer after encrypting with
// this Cipher.
func (e *CBK) Write(w io.Writer, b []byte) (int, error) {
	if e.buf == nil {
		e.buf = make([]byte, size+1)
	} else if e.total == -1 {
		e.total = len(e.buf) - 1
	}
	var n, i int
	for n < len(b) {
		if e.pos >= e.total {
			if _, err := e.flushOutput(w); err != nil {
				return n, err
			}
		}
		i = copy(e.buf[e.pos:e.total], b[n:])
		e.pos += i
		n += i
	}
	if e.pos < e.total {
		return n, nil
	}
	o, err := e.flushOutput(w)
	if o < e.total {
		return n - (e.total - o), err
	}
	return n, err
}
