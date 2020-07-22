package util

import (
	"math/rand"

	// Import unsafe to use faster "cputicks" function instead of "time.Now().UnixNano()"
	_ "unsafe"
)

const (
	max        = 63 / size
	size       = 6
	mask       = 1<<size - 1
	alpha      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	lenUpper   = 52
	lenLower   = 26
	lenNumbers = 78
)

// Rand is the custom Random number generator, based on the current time as a seed.
var Rand = &random{Rand: rand.New(rand.NewSource(cputicks()))}

var (
	// All represents the string instruction set that contains all alpha-numeric characters.
	All set = [2]int{0, 62}
	// Upper represents the string instruction set that contains only uppercase non-numeric characters.
	Upper set = [2]int{26, 52}
	// Lower represents the string instruction set that contains only lowercase non-numeric characters.
	Lower set = [2]int{0, 26}
	// Numbers represents the string instruction set that contains only numeric characters.
	Numbers set = [2]int{52, 62}
	// Characters represents the string instruction set that contains mixed case non-numeric characters.
	Characters set = [2]int{0, 52}
)

type set [2]int
type random struct {
	*rand.Rand
}

//go:linkname cputicks runtime.cputicks
func cputicks() int64

// FastRand is a fast thread local random function. This should be used in place instead of 'Rand.Uint32()'.
//
// Taken from https://github.com/dgraph-io/ristretto/blob/master/z/rtutil.go Thanks!
//go:linkname FastRand runtime.fastrand
func FastRand() uint32

// FastRandN is a fast thread local random function. This should be used in place instead of 'Rand.Uint32n()'.
// This function will take a max value to specify.
func FastRandN(n int) uint32 {
	return FastRand() % uint32(n)
}
func (r *random) String(n int) string {
	return r.StringEx(All, n, n)
}
func (r *random) StringLower(n int) string {
	return r.StringEx(Lower, n, n)
}
func (r *random) StringUpper(n int) string {
	return r.StringEx(Upper, n, n)
}
func (r *random) StringNumber(n int) string {
	return r.StringEx(Numbers, n, n)
}
func (r *random) StringRange(m, x int) string {
	return r.StringEx(All, m, x)
}
func (r *random) StringCharacters(n int) string {
	return r.StringEx(Characters, n, n)
}
func (r *random) StringEx(t set, m, x int) string {
	if m < 0 || x <= 0 || m > x || t[0] > All[1] || t[1] > All[1] || t[0] >= t[1] {
		return ""
	}
	var n int
	if m == x {
		n = m
	} else {
		n = m + r.Intn(x)
	}
	s := make([]byte, n)
	for i, c, x := n-1, r.Int63(), max; i >= 0; {
		if x == 0 {
			c, x = r.Int63(), max
		}
		if d := int(c & mask); d < len(alpha) && d < t[1] && d > t[0] {
			s[i] = alpha[d]
			i--
		}
		c >>= size
		x--
	}
	return string(s)
}
func (r *random) StringLowerRange(m, x int) string {
	return r.StringEx(Lower, m, x)
}
func (r *random) StringUpperRange(m, x int) string {
	return r.StringEx(Upper, m, x)
}
func (r *random) StringNumberRange(m, x int) string {
	return r.StringEx(Numbers, m, x)
}
func (r *random) StringCharactersRange(m, x int) string {
	return r.StringEx(Characters, m, x)
}
