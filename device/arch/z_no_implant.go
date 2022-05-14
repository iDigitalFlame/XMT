//go:build !implant

package arch

// String returns the name of the Architecture type.
func (a Architecture) String() string {
	switch a {
	case X86:
		return "32bit"
	case X64:
		return "64bit"
	case ARM:
		return "ARM"
	case WASM:
		return "WASM"
	case Risc:
		return "RiscV"
	case Mips:
		return "MIPS"
	case ARM64:
		return "ARM64"
	case PowerPC:
		return "PowerPC"
	}
	return "Unknown"
}
