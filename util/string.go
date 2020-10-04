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

// FastUTF8Match is a function that will return true if both of the strings match regardless of case (case insensitive).
// This function ONLY works properly on UTF8 characters as the tradeoff for fastness.
func FastUTF8Match(s, m string) bool {
	if len(s) != len(m) {
		return false
	}
	for i := range s {
		switch {
		case s[i] == m[i]:
		case m[i] > 96 && s[i]+32 == m[i]:
		case s[i] > 96 && m[i]+32 == s[i]:
		default:
			return false
		}
	}
	return true
}
