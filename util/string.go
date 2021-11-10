package util

const table = "0123456789ABCDEF"

// Decode is used to un-encode a string written in a XOR byte array "encrypted" by the specified key.
// This function returns the string value of the result but also modifies the input array, which can
// be used to re-use the resulting string.
func Decode(k, d []byte) string {
	if len(k) == 0 || len(d) == 0 {
		return ""
	}
	for i := 0; i < len(d); i++ {
		d[i] = d[i] ^ k[i%len(k)]
	}
	return string(d)
}

// ByteHexString is a simple function that will quickly lookup a byte value to it's associated hex value.
func ByteHexString(b byte) string {
	if b < 16 {
		return table[b&0x0F : (b&0x0F)+1]
	}
	return table[b>>4:(b>>4)+1] + table[b&0x0F:(b&0x0F)+1]
}
