// Package text is a simple package for generating random string values with
// complex requirements and regular expressions.
//
// This package exposes a 'Rand' struct that can be used to generate multiple
// types of string values.
//
// The other exported types allow for generation of mutable expressions that can
// be used to generate matching regular expression values. These work well with
// any package that works with stringers, such as the "wc2" package.
//
package text

import (
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util"
)

const (
	max        = 63 / size
	size       = 6
	mask       = 1<<size - 1
	lenUpper   = 52
	lenLower   = 26
	lenNumbers = 78
)

// Rand is the custom Random number generator, based on the 'util.Rand' choice.
// This random can be used to create custom random string values.
var Rand = &random{s: util.Rand}

var (
	// All represents the string instruction set that contains all alpha-numeric
	// characters.
	All set = [2]byte{0, 62}
	// Upper represents the string instruction set that contains only uppercase
	// non-numeric characters.
	Upper set = [2]byte{26, 52}
	// Lower represents the string instruction set that contains only lowercase
	// non-numeric characters.
	Lower set = [2]byte{0, 26}
	// Numbers represents the string instruction set that contains only numeric
	// characters.
	Numbers set = [2]byte{52, 62}
	// Characters represents the string instruction set that contains mixed case
	// non-numeric characters.
	Characters set = [2]byte{0, 52}
)

type set [2]byte
type random struct {
	s interface {
		Intn(int) int
		Int63() int64
	}
}

func (s set) String(n int) string {
	return Rand.StringEx(s, n, n)
}
func (r random) String(n int) string {
	return r.StringEx(All, n, n)
}
func (r random) StringLower(n int) string {
	return r.StringEx(Lower, n, n)
}
func (r random) StringUpper(n int) string {
	return r.StringEx(Upper, n, n)
}
func (r random) StringNumber(n int) string {
	return r.StringEx(Numbers, n, n)
}
func (r random) StringRange(m, x int) string {
	return r.StringEx(All, m, x)
}
func (r random) StringCharacters(n int) string {
	return r.StringEx(Characters, n, n)
}
func (r random) StringEx(t set, m, x int) string {
	if m < 0 || x <= 0 || m > x || t[0] > All[1] || t[1] > All[1] || t[0] >= t[1] {
		return ""
	}
	var n int
	if m == x {
		n = m
	} else {
		n = m + r.s.Intn(x)
	}
	if n > data.MaxSlice {
		n = data.MaxSlice
	}
	s := make([]byte, n)
	for i, c, x := n-1, r.s.Int63(), max; i >= 0; {
		if x == 0 {
			c, x = r.s.Int63(), max
		}
		if d := int(c & mask); d < len(alpha) && d < int(t[1]) && d > int(t[0]) {
			s[i] = alpha[d]
			i--
		}
		c >>= size
		x--
	}
	return string(s)
}
func (r random) StringLowerRange(m, x int) string {
	return r.StringEx(Lower, m, x)
}
func (r random) StringUpperRange(m, x int) string {
	return r.StringEx(Upper, m, x)
}
func (r random) StringNumberRange(m, x int) string {
	return r.StringEx(Numbers, m, x)
}
func (r random) StringCharactersRange(m, x int) string {
	return r.StringEx(Characters, m, x)
}
