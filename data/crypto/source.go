package crypto

import (
	"crypto/sha512"
	"math/rand"
	"time"
)

const defaultSource int64 = 0x123456789FABCD

// MultiSource is a struct that is a random Source that can use multiple source providers and spreads
// the calls among them in a random manner.
type MultiSource struct {
	rng  *rand.Rand
	s    []rand.Source64
	seed int64
}
type stringer interface {
	String() string
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

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (m *MultiSource) Uint64() uint64 {
	if len(m.s) == 0 {
		return m.rng.Uint64()
	}
	return m.s[m.rng.Intn(len(m.s))].Uint64()
}

// Add will append the Source values to this MultiSource instance.
func (m *MultiSource) Add(s ...rand.Source64) {
	if m.s == nil {
		m.s = make([]rand.Source64, 0, len(s))
	}
	for i := range s {
		m.seed += s[i].Int63()
		m.s = append(m.s, s[i])
	}
	m.rng.Seed(m.seed)
}

// NewSource creates a random Source from the provided interface.
// This function supports all types of Golang primitives.
func NewSource(seed interface{}) rand.Source64 {
	return NewSourceEx(4, seed)
}

// NewMultiSource creates a new MultiSource struct instance.
func NewMultiSource(s ...rand.Source64) *MultiSource {
	m := &MultiSource{rng: rand.New(rand.NewSource(defaultSource))}
	if len(s) > 0 {
		m.Add(s...)
	}
	return m
}

// NewSourceEx creates a random Source from the provided interface.
// This function supports all types of Golang primitives. This function
// allows for supplying the rounds value, which defaults to the value of 4.
func NewSourceEx(rounds int, seed interface{}) rand.Source64 {
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
		h := sha512.New()
		switch x := seed.(type) {
		case []byte:
			h.Write(x)
		case string:
			h.Write([]byte(x))
		case stringer:
			h.Write([]byte(x.String()))
		default:
			// b = []byte(fmt.Sprintf("%s", seed))
			// NOTE(dij):
			//    Removing the need for "fmt" here.
			//    I don't think this is used enough to
			//    Adding as static instead.
			h.Write([]byte("[invalid]"))
		}
		for _, v := range h.Sum(nil) {
			s += int64(v)
		}
		h = nil
	}
	if s == 0 {
		return rand.NewSource(defaultSource).(rand.Source64)
	}
	for x := 0; x < rounds; x++ {
		s = s + (s * 2)
	}
	// NOTE(dij): The underlying type here is '*rand.rngSource' which supports Source64.
	//            If this panics, there is some else extremely wrong.
	return rand.NewSource(s).(rand.Source64)
}
