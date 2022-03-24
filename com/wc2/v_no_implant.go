//go:build !implant

package wc2

func (l *listener) String() string {
	return "WC2/" + l.Server.Addr
}
