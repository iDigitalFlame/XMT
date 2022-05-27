//go:build windows && crypt

package pipe

import "github.com/iDigitalFlame/xmt/util/crypt"

var (
	// PermEveryone is the SDDL string used in Windows Pipes to allow anyone to
	// write and read to the listening Pipe
	//
	// This can be used for Pipe communcation between privilege boundaries.
	//
	// Can be applied with the ListenPerm function.
	PermEveryone = crypt.Get(32) // D:PAI(A;;FA;;;WD)(A;;FA;;;SY)

	// ErrTimeout is an error returned by the 'Dial*' functions when the
	// specified timeout was reached when attempting to connect to a Pipe.
	ErrTimeout = &errno{m: crypt.Get(33), t: true} // connection timeout
	// ErrEmptyConn is an error received when the 'Listen' function receives a
	// shortly lived Pipe connection.
	//
	// This error is only temporary and does not indicate any Pipe server failures.
	ErrEmptyConn = &errno{m: crypt.Get(34), t: true} // empty connection
)

// Format will ensure the path for this Pipe socket fits the proper OS based
// pathname. Valid pathnames will be returned without any changes.
func Format(s string) string {
	if len(s) > 2 && s[0] == '\\' && s[1] == '\\' {
		return s
	}
	return crypt.Get(35) + s // \\.\pipe\
}
