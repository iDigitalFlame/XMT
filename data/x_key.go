//go:build !nokeyset

package data

// Crypt will perform an "encryption" operation on the underlying Chunk buffer.
// No bytes are added or removed and this will not change the Chunk's size.
//
// If the Chunk is empty, 'nokeyset' was specified on build or the Key is nil,
// this is a NOP.
func (c *Chunk) Crypt(k *Key) {
	if len(c.buf) == 0 || k == nil {
		return
	}
	for i := range c.buf {
		c.buf[i] = c.buf[i] ^ (*k)[i%KeySize]
	}
}
