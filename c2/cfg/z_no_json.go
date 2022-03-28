//go:build nojson || implant

package cfg

import "github.com/iDigitalFlame/xmt/util/xerr"

func (cBit) String() string {
	return ""
}

// String returns a string representation of the data included in this Config
// instance. Each separate setting will be seperated by commas.
func (Config) String() string {
	return ""
}

// JSON will combine the supplied settings into a JSON payload and returned in
// a byte slice. This will return any validation errors during conversion.
//
// Not valid when the 'binonly' tag is specified.
func JSON(_ ...Setting) ([]byte, error) {
	return nil, xerr.Sub("json disabled", 0x2)
}

// MarshalJSON will attempt to convert the raw binary data in this Config
// instance into a JSON formart.
//
// The only error that may occur is 'ErrInvalidSetting' if an invalid
// setting or data value is encountered during conversion.
//
// Not valid when the 'binonly' tag is specified.
func (Config) MarshalJSON() ([]byte, error) {
	return nil, xerr.Sub("json disabled", 0x2)
}

// UnmarshalJSON will attempt to convert the JSON data provided into this Config
// instance.
//
// Errors during parsing or formatting will be returned along with the
// 'ErrInvalidSetting' error if parsed data contains invalid values.
//
// Not valid when the 'binonly' tag is specified.
func (Config) UnmarshalJSON(_ []byte) error {
	return xerr.Sub("json disabled", 0x2)
}
