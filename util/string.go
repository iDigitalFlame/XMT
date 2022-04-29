package util

// Decode is used to un-encode a string written in a XOR byte array "encrypted"
// by the specified key.
//
// This function returns the string value of the result but also modifies the
// input array, which can be used to re-use the resulting string.
//
// NOTE(dij): Is this still used?
func Decode(k, d []byte) string {
	if len(k) == 0 || len(d) == 0 {
		return ""
	}
	for i := range d {
		d[i] = d[i] ^ k[i%len(k)]
	}
	return string(d)
}
