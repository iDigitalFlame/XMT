package registry

import "github.com/iDigitalFlame/xmt/util/xerr"

// Registry value types.
const (
	TypeString       = 1
	TypeExpandString = 2
	TypeBinary       = 3
	TypeDword        = 4
	TypeStringList   = 7
	TypeQword        = 11
)

var (
	// ErrUnexpectedSize is returned when the key data size was unexpected.
	ErrUnexpectedSize = xerr.Sub("unexpected key size", 0x10)
	// ErrUnexpectedType is returned by Get*Value when the value's type was
	// unexpected.
	ErrUnexpectedType = xerr.Sub("unexpected key type", 0xD)
)
