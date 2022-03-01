//go:build implant
// +build implant

package xerr

var (
	// Concat is a compile time constant to help signal if complex string values
	// should be concatinated inline.
	//
	// This helps prevent debugging when the "-tags implant" option is enabled.
	Concat = false

	table = "0123456789ABCDEF"
)

type numErr uint16

// New creates a new string backed error interface and returns it.
// This error struct does not support Unwrapping.
//
// The resulting errors created will be comparable.
func New(_ string) error {
	return &err{}
}
func (e numErr) Error() string {
	if e < 0xFF {
		return "0x" + byteHexStr(byte(e))
	}
	return "0x" + byteHexStr(byte(e>>8)) + byteHexStr(byte(e))
}
func byteHexStr(b byte) string {
	if b == 0 {
		return "0"
	}
	if b < 16 {
		return table[b&0x0F : (b&0x0F)+1]
	}
	return table[b>>4:(b>>4)+1] + table[b&0x0F:(b&0x0F)+1]
}
func (e numErr) String() string {
	return e.Error()
}

// Wrap creates a new error that wraps the specified error.
//
// If not nil, this function will append ": " + 'Error()' to the resulting
// string message.
func Wrap(_ string, e error) error {
	if e == nil {
		return &err{}
	}
	return e
}

// Sub creates a new string backed error interface and returns it.
// This error struct does not support Unwrapping.
//
// If the "-tags implant" option is selected, the second value, the error code,
// will be used instead, otherwise it's ignored.
//
// The resulting errors created will be comparable.
func Sub(_ string, c uint16) error {
	return numErr(c)
}
