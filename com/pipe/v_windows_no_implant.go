//go:build windows && !implant
// +build windows,!implant

package pipe

// String returns a string representation of this listener.
func (l *Listener) String() string {
	return "PIPE/" + string(l.addr)
}
