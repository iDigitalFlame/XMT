package crypto

import (
	crypto "crypto/rand"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sync"
)

const (
	size    = 16
	sizeMax = 128
)

var (
	chains = sync.Pool{
		New: func() interface{} {
			b := make([]byte, size+1)
			return &b
		},
	}
	tables = sync.Pool{
		New: func() interface{} {
			b := make([][]byte, size+1)
			for i := 0; i < len(b); i++ {
				b[i] = make([]byte, 256)
			}
			return &b
		},
	}
)

// CBK is the representation of the CBK Cipher.
// CBK is a block based cipher that allows for a variable size index in encoding.
type CBK struct {
	A, B, C, D byte
	Source     rand.Source

	buf        []byte
	index      uint8
	pos, total int
}

// NewCBK returns a new CBK Cipher with the D value specified. The other A, B and C values
// are randomally generated at runtime.
func NewCBK(d int) *CBK {
	c, _ := NewCBKEx(d, size, nil)
	return c
}

// Reset resets the encryption keys and sets them to new random bytes.
func (e *CBK) Reset() error {
	if _, err := crypto.Read(e.buf[0:3]); err != nil {
		return fmt.Errorf("unable to generate random values: %w", err)
	}
	e.A = e.buf[0]
	e.B = e.buf[1]
	e.C = e.buf[2]
	e.pos, e.index = 0, 0
	return nil
}

// BlockSize returns the cipher's block BlockSize.
func (e CBK) BlockSize() int {
	return len(e.buf) - 1
}

// Shuffle will switch around the bytes in the array based on the Cipher bytes.
func (e CBK) Shuffle(b []byte) {
	if len(b) > 1 {
		b[0] += e.A
	}
	for i := byte(0); i < byte(len(b)); i++ {
		switch {
		case i%e.A == 0:
			b[i] += (e.A - i)
		case e.C%i == 0:
			b[i] += (e.B - e.D)
		case i == e.D:
			b[i] -= (e.A + i)
		default:
			if i%2 == 0 {
				b[i] += e.B / 3
			} else {
				b[i] += e.C / 5
			}
		}
	}
}
func clear(b *[]byte, z *[][]byte) {
	for i := 0; i < len(*b); i++ {
		(*b)[i] = 0
	}
	chains.Put(b)
	if z != nil {
		tables.Put(z)
	}
}

// Deshuffle will reverse the switch around the bytes in the array based on the Cipher bytes.
func (e CBK) Deshuffle(b []byte) {
	if len(b) > 1 {
		b[0] -= e.A
	}
	for i := byte(0); i < byte(len(b)); i++ {
		switch {
		case i%e.A == 0:
			b[i] -= (e.A - i)
		case e.C%i == 0:
			b[i] -= (e.B - e.D)
		case i == e.D:
			b[i] += (e.A + i)
		default:
			if i%2 == 0 {
				b[i] -= e.B / 3
			} else {
				b[i] -= e.C / 5
			}
		}
	}
}
func (e *CBK) cipherTable(b []byte) {
	b[0] = byte(uint16(e.index+1)*uint16(e.D+1) + e.adjust(uint16(e.D)))
	for i := byte(1); i < byte(len(b))-1; i++ {
		switch {
		case i <= 6:
			if i%2 == 0 {
				b[i] = byte(uint16(e.index) - uint16(e.A) + uint16(e.B-(i-e.C)) + uint16(i) - e.adjust(uint16(e.A)))
			} else {
				b[i] = byte(uint16(e.index) - uint16(e.A) + uint16(e.B-(i-3)) + uint16(i) - e.adjust(uint16(e.A)))
			}
		case i > 6 && i <= 11:
			b[i] = byte(uint16(e.C) - uint16(e.B) + uint16((e.index+1)*i) + e.adjust(uint16(e.C)))
		case i > 11:
			b[i] = byte(e.adjust(uint16(e.B+e.C)) + uint16(e.D) - uint16(len(b)-1) - uint16(e.D) + uint16(e.A-e.C))
		}
	}
	b[len(b)-1] = byte(e.adjust(uint16(e.B+e.C)) + uint16(e.index) - uint16(len(b)-1) - uint16(e.D) + uint16(e.A-e.C))
}
func (e *CBK) adjust(i uint16) uint16 {
	if e.Source != nil {
		return uint16(e.Source.Int63() * int64(i+1))
	}
	return uint16(math.Max(float64((e.A^e.B)-e.C)*float64(i+1), 1))
}

// Encrypt encrypts the first block in src into dst. Dst and src must overlap entirely or not at all.
func (e *CBK) Encrypt(dst, src []byte) {
	copy(dst, src)
	e.Shuffle(dst)
	e.scramble(dst, true)
}

// Decrypt decrypts the first block in src into dst. Dst and src must overlap entirely or not at all.
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
		o    = *chains.Get().(*[]byte)
		x    = e.adjust(uint16(e.A*e.B) + uint16(e.D))
		y    = e.adjust(uint16((e.C-e.D)*e.A) + x + e.adjust(uint16(e.index)))
		z    = e.adjust(uint16(byte(x*y) + e.B - byte(e.D*e.index)))
		i    int8
		g, h byte
	)
	if d {
		i = 5
	}
	for (i < 6 && !d) || (i >= 0 && d) {
		g = byte(math.Abs(float64((byte(z*y) + e.blockIndex(true, uint16(uint16(e.D*e.A)+uint16(i)+x), uint16(uint16(e.D)+uint16(e.index)))) % 8)))
		h = byte(math.Abs(float64((byte(y) - e.blockIndex(false, uint16(y+uint16(e.D)+uint16(e.index*uint8(i+1))), uint16(uint16(e.D)+x+uint16(byte(uint16(i)*z)*e.A)))) % 8)))
		if g != h {
			if !d {
				b[h], b[g] = byte(((((b[h]>>4)&0xF)&0xF)<<4)|(b[g]&0xF)), byte(((((b[g]>>4)&0xF)&0xF)<<4)|(b[h]&0xF))
				b[g+1], b[h+1] = byte(((((b[g+1]>>4)&0xF)&0xF)<<4)|(b[h+1]&0xF)), byte(((((b[h+1]>>4)&0xF)&0xF)<<4)|(b[g+1]&0xF))
			} else {
				b[h], b[g] = byte(((((b[h]>>4)&0xF)&0xF)<<4)|(b[g]&0xF)), byte(((((b[g]>>4)&0xF)&0xF)<<4)|(b[h]&0xF))
				b[g+1], b[h+1] = byte(((((b[g+1]>>4)&0xF)&0xF)<<4)|(b[h+1]&0xF)), byte(((((b[h+1]>>4)&0xF)&0xF)<<4)|(b[g+1]&0xF))
			}
			copy(o[0:2], b[g*2:g*2+2])
			copy(b[g*2:], b[h*2:h*2+2])
			copy(b[h*2:], o[0:2])
			if d {
				b[h], b[g] = byte(((((b[h]>>4)&0xF)&0xF)<<4)|(b[g]&0xF)), byte(((((b[g]>>4)&0xF)&0xF)<<4)|(b[h]&0xF))
				b[g+1], b[h+1] = byte(((((b[g+1]>>4)&0xF)&0xF)<<4)|(b[h+1]&0xF)), byte(((((b[h+1]>>4)&0xF)&0xF)<<4)|(b[g+1]&0xF))
			} else {
				b[h], b[g] = byte(((((b[h]>>4)&0xF)&0xF)<<4)|(b[g]&0xF)), byte(((((b[g]>>4)&0xF)&0xF)<<4)|(b[h]&0xF))
				b[g+1], b[h+1] = byte(((((b[g+1]>>4)&0xF)&0xF)<<4)|(b[h+1]&0xF)), byte(((((b[h+1]>>4)&0xF)&0xF)<<4)|(b[g+1]&0xF))
			}
		}
		if d {
			i--
		} else {
			i++
		}
	}
	clear(&o, nil)
}
func (e *CBK) readInput(r io.Reader) (int, error) {
	n, err := r.Read(e.buf)
	if err != nil {
		e.total = 0
		return 0, err
	}
	if n <= 0 {
		e.total = 0
		return 0, io.EOF
	}
	e.index++
	if e.index > 30 {
		e.index = 0
	}
	var (
		t = *chains.Get().(*[]byte)
		c = *tables.Get().(*[][]byte)
	)
	e.cipherTable(t)
	e.Deshuffle(e.buf)
	e.scramble(e.buf, true)
	for x := 0; x < len(c); x++ {
		for z := 0; z < len(c[x]); z++ {
			c[x][t[x]&0xFF] = byte(z)
			t[x]++
		}
	}
	for i := 0; i < len(e.buf); i++ {
		e.buf[i] = c[i&0xF][e.buf[i]&0xFF]
	}
	e.total = int(e.buf[len(e.buf)-1])
	e.pos = 0
	clear(&t, &c)
	if e.total == 0 {
		return 0, io.EOF
	}
	return n, nil
}
func (e CBK) blockIndex(a bool, t, i uint16) byte {
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
		return byte(((((9 + i + uint16(e.A*e.D)) / 4) + (t / 2) + (2*i + 1 + uint16(e.D))) / (((i + 3) / (5 + t)) + 6)))
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
	e.index++
	if e.index > 30 {
		e.index = 0
	}
	e.buf[e.total] = byte(e.pos)
	var (
		t = *chains.Get().(*[]byte)
		c = *tables.Get().(*[][]byte)
	)
	e.cipherTable(t)
	for x := 0; x < len(c); x++ {
		for z := 0; z < len(c[x]); z++ {
			c[x][z] = t[x]
			t[x]++
		}
	}
	for i := 0; i < len(e.buf); i++ {
		e.buf[i] = c[i&0xF][e.buf[i]&0xFF]
	}
	e.scramble(e.buf, false)
	e.Shuffle(e.buf)
	n, err := w.Write(e.buf)
	clear(&t, &c)
	if err != nil {
		return 0, err
	}
	e.pos = 0
	return n, nil
}

// Read reads the contents of the Reader to the byte array after decrypting with this Cipher.
func (e *CBK) Read(r io.Reader, b []byte) (int, error) {
	if e.buf == nil {
		e.buf = make([]byte, size+1)
	}
	if e.total-e.pos > len(b) {
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
		e.pos += i
		if e.pos >= e.total && e.total >= len(e.buf)-1 {
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

// Write writes the contents of the byte array to the Writer after encrypting with this Cipher.
func (e *CBK) Write(w io.Writer, b []byte) (int, error) {
	if e.buf == nil {
		e.buf = make([]byte, size+1)
	} else if e.total == -1 {
		e.total = len(e.buf) - 1
	}
	var n, i int
	for n < len(b) {
		if e.pos >= e.total {
			_, err := e.flushOutput(w)
			if err != nil {
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

// NewCBKEx returns a new CBK Cipher with the D value, BlockSize and Entropy source specified. The other
// A, B and C values are randomally generated at runtime.
func NewCBKEx(d int, size int, source rand.Source) (*CBK, error) {
	if size < 0 || size > sizeMax || math.Floor(math.Log2(float64(size))) != math.Ceil(math.Log2(float64(size))) {
		return nil, fmt.Errorf("block size %d must be between 16 and 128 and a power of two", size)
	}
	c := &CBK{D: byte(d), buf: make([]byte, size+1), total: -1, Source: source}
	c.Reset()
	return c, nil
}
