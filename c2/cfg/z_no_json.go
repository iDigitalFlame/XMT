//go:build binonly
// +build binonly

package cfg

func (cBit) String() string {
	return ""
}
func bitFromName(_ string) cBit {
	return invalid
}

// JSON will combine the supplied settings into a JSON payload and returned in
// a byte slice. This will return any validation errors during conversion.
//
// Not valid when the 'binonly' tag is specified.
func JSON(_ ...Setting) ([]byte, error) {
	return nil, ErrInvalidSetting
}

// MarshalJSON will attempt to convert the raw binary data in this Config
// instance into a JSON formart.
//
// The only error that may occur is 'ErrInvalidSetting' if an invalid
// setting or data value is encountered during conversion.
func (Config) MarshalJSON() ([]byte, error) {
	return nil, ErrInvalidSetting
}

// UnmarshalJSON will attempt to convert the JSON data provided into this Config
// instance.
//
// Errors during parsing or formatting will be returned along with the
// 'ErrInvalidSetting' error if parsed data contains invalid values.
func (Config) UnmarshalJSON(_ []byte) error {
	return ErrInvalidSetting
}
