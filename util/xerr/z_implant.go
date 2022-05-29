//go:build implant

package xerr

const (
	// ExtendedInfo is a compile time constant to help signal if complex string
	// values should be concatinated inline.
	//
	// This helps prevent debugging when the "-tags implant" option is enabled.
	ExtendedInfo = false

	table = "0123456789ABCDEF"
)

type numErr uint8

func (e numErr) Error() string {
	return "0x" + byteHexStr(e)
}
func byteHexStr(b numErr) string {
	if b == 0 {
		return "0"
	}
	if b < 16 {
		return table[b&0x0F : (b&0x0F)+1]
	}
	return table[b>>4:(b>>4)+1] + table[b&0x0F:(b&0x0F)+1]
}

// Sub creates a new string backed error interface and returns it.
// This error struct does not support Unwrapping.
//
// If the "-tags implant" option is selected, the second value, the error code,
// will be used instead, otherwise it's ignored.
//
// The resulting errors created will be comparable.
func Sub(_ string, c uint8) error {
	return numErr(c)
}

// Wrap creates a new error that wraps the specified error.
//
// If not nil, this function will append ": " + 'Error()' to the resulting
// string message and will keep the original error for unwrapping.
//
// If "-tags implant" is specified, this will instead return the wrapped error
// directly.
func Wrap(_ string, e error) error {
	return e
}
