//go:build windows && !crypt

package pipe

// PermEveryone is the SDDL string used in Windows Pipes to allow anyone to
// write and read to the listening Pipe
//
// This can be used for Pipe communcation between privilege boundaries.
//
// Can be applied with the ListenPerm function.
const PermEveryone = "D:PAI(A;;FA;;;WD)(A;;FA;;;SY)"

var (
	// ErrTimeout is an error returned by the 'Dial*' functions when the
	// specified timeout was reached when attempting to connect to a Pipe.
	ErrTimeout = &errno{m: "connection timeout", t: true}
	// ErrEmptyConn is an error received when the 'Listen' function receives a
	// shortly lived Pipe connection.
	//
	// This error is only temporary and does not indicate any Pipe server failures.
	ErrEmptyConn = &errno{m: "empty connection", t: true}
)

// Format will ensure the path for this Pipe socket fits the proper OS based
// pathname. Valid pathnames will be returned without any changes.
func Format(s string) string {
	if len(s) > 2 && s[0] == '\\' && s[1] == '\\' {
		return s
	}
	return `\\.\pipe` + "\\" + s
}
