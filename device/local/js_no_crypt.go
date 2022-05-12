//go:build (wasm || js) && !crypt

package local

func sysID() []byte {
	return nil
}
func version() string {
	return "JavaScript"
}
func isElevated() uint8 {
	return 0
}
