//go:build !crypt

package device

const emptyIP = "0.0.0.0"

func (d deviceOS) String() string {
	switch d {
	case Windows:
		return "Windows"
	case Linux:
		return "Linux"
	case Unix:
		return "Unix/BSD"
	case Mac:
		return "MacOS"
	case Unsupported:
		return "Unsupported"
	}
	return "Unknown"
}
func (d deviceArch) String() string {
	switch d {
	case Arch86:
		return "32bit"
	case Arch64:
		return "64bit"
	case ArchARM:
		return "ARM"
	case ArchWASM:
		return "WASM"
	case ArchRisc:
		return "RiscV"
	case ArchMips:
		return "MIPS"
	case ArchARM64:
		return "ARM64"
	case ArchPowerPC:
		return "PowerPC"
	}
	return "Unknown"
}
