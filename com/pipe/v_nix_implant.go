//go:build !windows && implant

package pipe

func (listener) String() string {
	return ""
}
