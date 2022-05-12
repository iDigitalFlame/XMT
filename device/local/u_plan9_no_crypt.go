//go:build plan9 && !crypt

package local

func uname() string {
	return "plan9"
}
