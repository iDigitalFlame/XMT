package util

import (
	"math/rand"
	"time"
)

const (
	randMax   = 63 / randSize
	randSize  = 6
	randMask  = 1<<randSize - 1
	randAlpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	randBytes = make([]byte, 127)
	// Random is the general random number generator.
	Random = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func init() {
	for i := range randBytes {
		randBytes[i] = byte(i)
	}
}

// RandInt returns a psueo-random Integer between 0 and "n".
func RandInt(n int) int {
	if n <= 0 {
		return 0
	}
	return Random.Intn(n)
}

// RandBytes will fill the supplied byte array 'b' with psueo-random bytes.
func RandBytes(b []byte) {
	for i, z, j := len(b)-1, Random.Int63(), randMax; i >= 0; {
		if j == 0 {
			z, j = Random.Int63(), randMax
		}
		if y := int(z & randMask); y < len(randBytes) {
			b[i] = randBytes[y]
			i--
		}
		z >>= randSize
		j--
	}
}

// RandInt32 returns a psueo-random 32bit Integer between 0 and "n".
func RandInt32(n int) int32 {
	if n <= 0 {
		return 0
	}
	return Random.Int31n(int32(n))
}

// RandString returns a psueo-random string of "n" length.
func RandString(n int64) string {
	if n <= 0 {
		return Empty
	}
	b := make([]byte, n)
	for i, c, r := n-1, Random.Int63(), randMax; i >= 0; {
		if r == 0 {
			c, r = Random.Int63(), randMax
		}
		if d := int(c & randMask); d < len(randAlpha) {
			b[i] = randAlpha[d]
			i--
		}
		c >>= randSize
		r--
	}
	return string(b)
}
