//go:build (wasm || js) && !crypt
// +build wasm js

package local

func sysID() []byte {
	return nil
}
func version() string {
	return "JavaScript"
}
func isElevated() bool {
	return false
}
