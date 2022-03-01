//go:build crypt
// +build crypt

package device

import (
	"regexp"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

var (
	home    = crypt.Get(17) // $HOME
	emptyIP = crypt.Get(18) // "0.0.0.0"

	envExp = regexp.MustCompile(crypt.Get(19)) // %[\w\d()_-]+%|\$[\w\d_-]+|\$\{[[\w\d_-]+\}
)

func (d deviceOS) String() string {
	switch d {
	case Windows:
		return crypt.Get(20) // Windows
	case Linux:
		return crypt.Get(21) // Linux
	case Unix:
		return crypt.Get(22) // Unix/BSD
	case Mac:
		return crypt.Get(23) // MacOS
	case Unsupported:
		return crypt.Get(24) // Unsupported
	}
	return crypt.Get(25) // Unknown
}
func (d deviceArch) String() string {
	switch d {
	case Arch86:
		return crypt.Get(26) // 32bit
	case Arch64:
		return crypt.Get(27) // 64bit
	case ArchARM:
		return crypt.Get(28) // ARM
	case ArchWASM:
		return crypt.Get(29) // WASM
	case ArchRisc:
		return crypt.Get(30) // RiscV
	case ArchMips:
		return crypt.Get(31) // MIPS
	case ArchARM64:
		return crypt.Get(32) // ARM64
	case ArchPowerPC:
		return crypt.Get(33) // PowerPC
	}
	return crypt.Get(25) // Unknown
}
