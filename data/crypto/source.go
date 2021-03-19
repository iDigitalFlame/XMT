package crypto

import (
	"crypto/sha512"
	"fmt"
	"hash"
	"math/rand"
	"sync"
	"time"
)

const defaultSource int64 = 0x123456789F

var hashers = sync.Pool{
	New: func() interface{} {
		return sha512.New()
	},
}

/*
// Source is an interface that supports seed assistance in Ciphers and other cryptographic functions.
type Source interface {
	Reset() error
	Next(uint16) uint16
}
*/

// MultiSource is a struct that is a random Source that can use multiple source providers and spreads
// the calls among them in a random manner.
type MultiSource struct {
	rng  *rand.Rand
	s    []rand.Source
	seed int64
}

// SHA512 returns the SHA512 value of the provided byte array. This function is not recommended for big arrays.
func SHA512(b []byte) []byte {
	h := hashers.Get().(hash.Hash)
	h.Write(b)
	r := h.Sum(nil)
	h.Reset()
	hashers.Put(h)
	return r
}

// Int63 returns a int64 number between zero and the max value.
func (m *MultiSource) Int63() int64 {
	if len(m.s) == 0 {
		return m.rng.Int63()
	}
	return m.s[m.rng.Intn(len(m.s))].Int63()
}

// Seed will set the seed value of this MultiSource instance.
func (m *MultiSource) Seed(n int64) {
	m.seed = n
	m.rng.Seed(n)
}

// Add will append the Source values to this MultiSource instance.
func (m *MultiSource) Add(s ...rand.Source) {
	if m.s == nil {
		m.s = make([]rand.Source, 0, len(s))
	}
	for i := range s {
		m.seed += s[i].Int63()
		m.s = append(m.s, s[i])
	}
	m.rng.Seed(m.seed)
}

// NewSource creates a random Source from the provided interface.
// This function supports all types of Golang primitives.
func NewSource(seed interface{}) rand.Source {
	return NewSourceEx(4, seed)
}

// NewMultiSource creates a new MultiSource struct instance.
func NewMultiSource(s ...rand.Source) *MultiSource {
	m := &MultiSource{rng: rand.New(rand.NewSource(defaultSource))}
	if len(s) > 0 {
		m.Add(s...)
	}
	return m
}

// NewSourceEx creates a random Source from the provided interface.
// This function supports all types of Golang primitives. This function
// allows for supplying the rounds value, which defaults to the value of 4.
func NewSourceEx(rounds int, seed interface{}) rand.Source {
	var s int64
	if rounds <= 0 {
		rounds = 1
	}
	switch i := seed.(type) {
	case int:
		s = int64(i)
	case bool:
		if i {
			s = 1
		}
	case uint:
		s = int64(i)
	case int8:
		s = int64(i)
	case uint8:
		s = int64(i)
	case int16:
		s = int64(i)
	case int32:
		s = int64(i)
	case int64:
		s = int64(i)
	case uint16:
		s = int64(i)
	case uint32:
		s = int64(i)
	case uint64:
		s = int64(i)
	case time.Time:
		s = i.Unix()
	case time.Duration:
		s = int64(i)
	default:
		var b []byte
		switch x := seed.(type) {
		case []byte:
			b = x
		case string:
			b = []byte(x)
		default:
			b = []byte(fmt.Sprintf("%s", seed))
		}
		for _, v := range SHA512(b) {
			s += int64(v)
		}
	}
	for x := 0; x < rounds; x++ {
		s = s + (s * 2)
	}
	return rand.NewSource(s)
}
