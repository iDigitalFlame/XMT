package crypto

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"time"
)

const (
	// DefaultRounds is the amount of rounds that the hash is passed through
	// when using the NewSource function.
	DefaultRounds = 10
	// DefaultSource is the default Source Seed used for MultiSource struct data
	// if a source is not provided during creation.
	DefaultSource int64 = 0x123456789F
)

// Source is an interface that supports seed assistance in Ciphers and other
// cryptographic functions.
type Source interface {
	Reset() error
	Next(uint16) uint16
}

// MultiSource is a struct that is a random Source that can use multiple
// source providers and spreads the calls among them in a random manner.
type MultiSource struct {
	s    []rand.Source
	rng  *rand.Rand
	seed int64
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
	return NewSourceEx(DefaultRounds, seed)
}

// NewMultiSource creates a new MultiSource struct instance.
func NewMultiSource(s ...rand.Source) *MultiSource {
	m := &MultiSource{rng: rand.New(rand.NewSource(DefaultSource))}
	if len(s) > 0 {
		m.Add(s...)
	}
	return m
}

// NewSourceEx creates a random Source from the provided interface.
// This function supports all types of Golang primitives. This function
// allows for supplying the rounds value, which defaults to the value of DefaultRounds.
func NewSourceEx(rounds int, seed interface{}) rand.Source {
	var s int64
	if rounds <= 0 {
		rounds = 1
	}
	switch seed.(type) {
	case int:
		s = int64(seed.(int))
	case bool:
		if seed.(bool) {
			s = 1
		}
	case uint:
		s = int64(seed.(uint))
	case int8:
		s = int64(seed.(int8))
	case uint8:
		s = int64(seed.(uint8))
	case int16:
		s = int64(seed.(int16))
	case int32:
		s = int64(seed.(int32))
	case int64:
		s = int64(seed.(int64))
	case uint16:
		s = int64(seed.(uint16))
	case uint32:
		s = int64(seed.(uint32))
	case uint64:
		s = int64(seed.(uint64))
	case time.Time:
		s = seed.(time.Time).Unix()
	case time.Duration:
		s = int64(seed.(time.Duration))
	default:
		var b []byte
		switch seed.(type) {
		case []byte:
			b = seed.([]byte)
		case string:
			b = []byte(seed.(string))
		default:
			b = []byte(fmt.Sprintf("%s", seed))
		}
		v := sha512.Sum512(b)
		s += int64(
			uint64(v[7]) | uint64(v[6])<<8 | uint64(v[5])<<16 | uint64(v[4])<<24 |
				uint64(v[3])<<32 | uint64(v[2])<<40 | uint64(v[1])<<48 | uint64(v[0])<<56,
		)
	}
	for x := 0; x < rounds; x++ {
		s = s + (s * 2)
	}
	return rand.NewSource(s)
}
