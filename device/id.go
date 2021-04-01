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

// ID is an alias for a byte array that represents a 32 byte client identification number. This is used for
// tracking and detection purposes. The first byte and the machine ID byte should NEVER be zero, otherwise it
// signals an invalid ID value or missing a random identifier.
type ID [IDSize]byte

func getID() ID {
	var (
		i      ID
		s, err = machineid.ProtectedID("xmtFramework-v2")
	)
	if err == nil {
		copy(i[:], s)
	} else {
		rand.Read(i[:])
	}
	rand.Read(i[MachineIDSize:])
	if i[0] == 0 {
		i[0] = 1
	}
	if i[MachineIDSize] == 0 {
		i[MachineIDSize] = 1
	}
	return i
}

// Empty returns true if this ID is considered empty.
func (i ID) Empty() bool {
	return i[0] == 0
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
	if i[MachineIDSize] == 0 {
		return i.string(0, MachineIDSize)
	}
	return i.string(MachineIDSize, IDSize)
}

// Seed will set the random portion of the ID value to the specified byte array value.
func (i *ID) Seed(b []byte) {
	if len(b) == 0 {
		return
	}
	copy(i[MachineIDSize:], b)
}

// Equal will return true if both ID values are equal in size and have the same Hash value.
func (i ID) Equal(a ID) bool {
	// The equals sign is a way faster method of comparing values.
	// return i.Hash() == a.Hash()
	return i == a
}

// Signature returns the signature portion of the ID value. This value is constant and unique for each device.
func (i ID) Signature() string {
	if i[MachineIDSize] == 0 {
		return i.string(0, MachineIDSize)
	}
	return i.string(0, MachineIDSize)
}

// FullString returns the full string representation of this ID instance.
func (i ID) FullString() string {
	return i.string(0, IDSize)
}

// Load will attempt to load the Session UUID from the specified file. This function will return an
// error if the file cannot be read or not found.
func (i ID) Load(s string) error {
	r, err := os.OpenFile(s, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	err = i.Read(r)
	if r.Close(); err != nil {
		return err
	}
	return nil
}

// Save will attempt to save the Session UUID to the specified file. This function will return an
// error if the file cannot be written to or created.
func (i ID) Save(s string) error {
	w, err := os.OpenFile(s, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	_, err = w.Write(i[:])
	if w.Close(); err != nil {
		return err
	}
	return nil
}

// Read will attempt to read up to 'IDSize' bytes from the reader into this ID.
func (i *ID) Read(r io.Reader) error {
	n, err := data.ReadFully(r, i[:])
	if err != nil {
		return err
	}
	if n != IDSize {
		return io.EOF
	}
	return nil
}

// Write will attempt to write the ID bytes into the supplied writer.
func (i ID) Write(w io.Writer) error {
	_, err := w.Write(i[:])
	return err
}

func (i ID) string(start, end int) string {
	s := builders.Get().(*strings.Builder)
	s.Grow(end - start)
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
	_, err := w.Write(i[:])
	return err
}

// UnmarshalStream transforms this struct from a binary format that is read from the supplied data.Reader.
func (i *ID) UnmarshalStream(r data.Reader) error {
	return i.Read(r)
}
