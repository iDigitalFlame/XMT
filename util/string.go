package util

import "github.com/iDigitalFlame/xmt/data/crypto"

// Decode is used to un-encode a string written in a XOR byte array "encrypted" by the specified key.
// This function returns the string value of the result but also modifies the input array, which can
// be used to re-use the resulting string.
func Decode(k, d []byte) string {
	if len(k) == 0 || len(d) == 0 {
		return ""
	}
	crypto.XOR(k).Operate(d)
	return string(d)
}
