//go:build implant && !noproxy

package c2

func (Proxy) prefix() string {
	return ""
}
func (*Session) name() string {
	return ""
}
