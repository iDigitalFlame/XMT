//go:build !ews || !implant

package c2

type container string

func (container) Wrap() {
}
func (container) Unwrap() {
}
func (c *container) Set(s string) {
	*c = container(s)
}
func (c container) String() string {
	return string(c)
}
