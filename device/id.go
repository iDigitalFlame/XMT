package device

import (
	"crypto/rand"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/denisbrodbeck/machineid"
	"github.com/iDigitalFlame/xmt/data"
)

const (
	// IDSize is the amount of bytes used to store the Host ID and
	// SessionID values.  The ID is the (HostID + SessionID).
	IDSize = 32
	// MachineIDSize is the amount of bytes that is used as the Host
	// specific ID value that does not change when on the same host.
	MachineIDSize = 28
)

const table = "0123456789ABCDEF"

var builders = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

// ID is an alias for a byte array that represents a 48 byte client identification number. This is used for
// tracking and detection purposes.
type ID []byte

func getID() ID {
	var (
		i      = ID(make([]byte, IDSize))
		s, err = machineid.ProtectedID("xmtFramework-v2")
	)
	if err == nil {
		copy(i, s)
	} else {
		rand.Read(i)
	}
	rand.Read(i[MachineIDSize:])
	return i
}

// Hash returns the 32bit hash sum of this ID value. The hash mechanism used is similar to the hash/fnv mechanism.
func (i ID) Hash() uint32 {
	h := uint32(2166136261)
	for x := range i {
		h *= 16777619
		h ^= uint32(i[x])
	}
	return h
}

// String returns a representation of this ID instance.
func (i ID) String() string {
	if len(i) < MachineIDSize {
		return i.string(0, len(i))
	}
	return i.string(MachineIDSize, len(i))
}

// Signature returns the signature portion of the ID value. This value is constant and unique for each device.
func (i ID) Signature() string {
	if len(i) < MachineIDSize {
		return i.string(0, len(i))
	}
	return i.string(0, MachineIDSize)
}

// FullString returns the full string representation of this ID instance.
func (i ID) FullString() string {
	return i.string(0, len(i))
}

// LoadSession will attempt to load the Session UUID from the specified file. This function will return an
// error if the file cannot be read or not found.
func LoadSession(s string) error {
	r, err := os.OpenFile(s, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	n, err := data.ReadFully(r, UUID)
	if r.Close(); err != nil {
		return err
	}
	if n != IDSize {
		return io.EOF
	}
	return nil
}

// SaveSession will attempt to save the Session UUID to the specified file. This function will return an
// error if the file cannot be written to or created.
func SaveSession(s string) error {
	w, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = w.Write(UUID)
	if w.Close(); err != nil {
		return err
	}
	return nil
}
func (i ID) string(start, end int) string {
	s := builders.Get().(*strings.Builder)
	for x := start; x < end; x++ {
		s.WriteByte(table[i[x]>>4])
		s.WriteByte(table[i[x]&0x0F])
	}
	r := s.String()
	s.Reset()
	builders.Put(s)
	return r
}

// MarshalStream transform this struct into a binary format and writes to the supplied data.Writer.
func (i ID) MarshalStream(w data.Writer) error {
	_, err := w.Write(i)
	return err
}

// UnmarshalStream transforms this struct from a binary format that is read from the supplied data.Reader.
func (i *ID) UnmarshalStream(r data.Reader) error {
	if len(*i) == 0 {
		*i = make([]byte, IDSize)
	}
	n, err := data.ReadFully(r, *i)
	if err != nil {
		return err
	}
	if n != IDSize {
		return io.EOF
	}
	return nil
}
