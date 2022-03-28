package regedit

import (
	"unsafe"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/device/winapi/registry"
)

// Entry represents a Windows registry entry.
//
// This may represent a Key or a Value. Values will also include their data in
// the 'Data' byte array and can be translated using any of the 'To*' functions
// to get the data repsented in it's proper type cast.
type Entry struct {
	Name string
	Data []byte
	Type uint32
}

// IsKey returns true if this entry represents a SubKey.
func (e Entry) IsKey() bool {
	return e.Type == 0
}

// IsZero returns true if this entry is an invalid Entry.
func (e Entry) IsZero() bool {
	return e.Type == 0 && len(e.Data) == 0 && len(e.Name) == 0
}

// ToBinary will attempt to return the data in the Data buffer as a binary
// array.
//
// This function returns an error if conversion fails or the specified type is
// not a BINARY type.
func (e Entry) ToBinary() ([]byte, error) {
	if e.Type != registry.TypeBinary {
		return nil, registry.ErrUnexpectedType
	}
	return e.Data, nil
}

// ToString will attempt to return the data in the Data buffer as a string.
//
// This function returns an error if conversion fails or the specified type is
// not a STRING or EXPAND_STRING type.
func (e Entry) ToString() (string, error) {
	if e.Type != registry.TypeString && e.Type != registry.TypeExpandString {
		return "", registry.ErrUnexpectedType
	}
	if len(e.Data) < 3 {
		return "", registry.ErrUnexpectedSize
	}
	return winapi.UTF16ToString((*[1 << 29]uint16)(unsafe.Pointer(&e.Data[0]))[: len(e.Data)/2 : len(e.Data)/2]), nil
}

// ToInteger will attempt to return the data in the Data buffer as an unsigned
// integer.
//
// This function returns an error if conversion fails or the specified type is
// not a DWORD or QWORD type.
func (e Entry) ToInteger() (uint64, error) {
	switch e.Type {
	case registry.TypeDword:
		if len(e.Data) != 4 {
			return 0, registry.ErrUnexpectedSize
		}
		_ = e.Data[3]
		return uint64(e.Data[0]) | uint64(e.Data[1])<<8 | uint64(e.Data[2])<<16 | uint64(e.Data[3])<<24, nil
	case registry.TypeQword:
		if len(e.Data) != 8 {
			return 0, registry.ErrUnexpectedSize
		}
		_ = e.Data[7]
		return uint64(e.Data[0]) | uint64(e.Data[1])<<8 | uint64(e.Data[2])<<16 | uint64(e.Data[3])<<24 |
			uint64(e.Data[4])<<32 | uint64(e.Data[5])<<40 | uint64(e.Data[6])<<48 | uint64(e.Data[7])<<56, nil
	}
	return 0, registry.ErrUnexpectedType
}

// ToStringList will attempt to return the data in the Data buffer as a string
// array.
//
// This function returns an error if conversion fails or the specified type is
// not a MULTI_STRING type.
func (e Entry) ToStringList() ([]string, error) {
	if e.Type != registry.TypeStringList {
		return nil, registry.ErrUnexpectedType
	}
	if len(e.Data) < 3 {
		return nil, registry.ErrUnexpectedSize
	}
	v := (*[1 << 29]uint16)(unsafe.Pointer(&e.Data[0]))[: len(e.Data)/2 : len(e.Data)/2]
	if len(v) == 0 {
		return nil, nil
	}
	if v[len(v)-1] == 0 {
		v = v[:len(v)-1]
	}
	r := make([]string, 0, len(v))
	for i, n := 0, 0; i < len(v); i++ {
		if v[i] > 0 {
			continue
		}
		r = append(r, string(winapi.UTF16Decode(v[n:i])))
		n = i + 1
	}
	return r, nil
}

// MarshalStream writes the data for this Entry to the supplied Writer.
func (e Entry) MarshalStream(w data.Writer) error {
	if err := w.WriteString(e.Name); err != nil {
		return err
	}
	if err := w.WriteUint32(e.Type); err != nil {
		return err
	}
	return w.WriteBytes(e.Data)
}

// UnmarshalStream reads the data for this Entry from the supplied Reader.
func (e *Entry) UnmarshalStream(r data.Reader) error {
	if err := r.ReadString(&e.Name); err != nil {
		return err
	}
	if err := r.ReadUint32(&e.Type); err != nil {
		return err
	}
	return r.ReadBytes(&e.Data)
}
