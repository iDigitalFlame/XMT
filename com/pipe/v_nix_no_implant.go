//go:build !windows && !implant

package pipe

func (l *listener) String() string {
	return "PIPE/" + l.Addr().String()
}
