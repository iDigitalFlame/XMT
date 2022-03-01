//go:build implant
// +build implant

package cmd

// DLLToASM will patch the DLL raw bytes and convert it into shellcode
// using thr SRDi launcher.
//   SRDi GitHub: https://github.com/monoxgas/sRDI
//
// The first string param is the function name which can be empty if not
// needed.
//
// The resulting byte slice can be used in an 'Asm' struct to directly load and
// run the DLL.
func DLLToASM(_ string, b []byte) []byte {
	return b
}
