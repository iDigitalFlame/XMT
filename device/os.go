package device

import (
	"os"
	"strings"
	"sync"
	"syscall"

	"github.com/iDigitalFlame/xmt/util"
)

const (
	// Windows represents the Windows family of Operating Systems.
	Windows deviceOS = 0x0
	// Linux represents the Linux family of Operating Systems
	Linux deviceOS = 0x1
	// Unix represents the Unix family of Operating Systems
	Unix deviceOS = 0x2
	// Mac represents the MacOS/BSD family of Operating Systems
	Mac deviceOS = 0x3
	// Unsupported represents a device type that does not have direct support
	// any may not work properly.
	Unsupported deviceOS = 0x4

	// Arch64 represents the 64-bit chipset family.
	Arch64 deviceArch = 0x0
	// Arch86 represents the 32-bit chipset family.
	Arch86 deviceArch = 0x1
	// ArchARM represents the ARM chipset family.
	ArchARM deviceArch = 0x2
	// ArchPowerPC represents the PowerPC chipset family.
	ArchPowerPC deviceArch = 0x3
	// ArchMips represents the MIPS chipset family.
	ArchMips deviceArch = 0x4
	// ArchRisc represents the RiscV chipset family.
	ArchRisc deviceArch = 0x5
	// ArchARM64 represents the ARM64 chipset family.
	ArchARM64 deviceArch = 0x6
	// ArchWASM represents the WASM/JavaScript software family.
	ArchWASM deviceArch = 0x7
	// ArchUnknown represents an unknown chipset family.
	ArchUnknown deviceArch = 0x8
)

var builders = sync.Pool{
	New: func() interface{} {
		return new(util.Builder)
	},
}

type deviceOS uint8
type deviceArch uint8

func init() {
	t := os.TempDir()
	syscall.Setenv("tmp", t)
	syscall.Setenv("temp", t)
}

// Expand attempts to determine environment variables from the current session
// and translate them from the supplied string.
//
// This function supports both Windows (%var%) and *nix ($var or ${var})
// variable substitutions.
func Expand(s string) string {
	if len(s) == 0 {
		return s
	}
	if len(s) >= 2 && s[0] == '~' && s[1] == '/' {
		// Account for shell expansion.
		s = home + s[1:]
	}
	v := envExp.FindAllStringIndex(s, -1)
	if len(v) == 0 {
		return s
	}
	b := builders.Get().(*util.Builder)
	b.Grow(len(s))
	b.WriteString(s[:v[0][0]])
	for i := range v {
		if i > 0 {
			b.WriteString(s[v[i-1][1]:v[i][0]])
		}
		a, x := 0, 0
		for ; s[v[i][0]+a] == '%' || s[v[i][0]+a] == '$' || s[v[i][0]+a] == '{'; a++ {
		}
		for ; s[v[i][1]-(1+x)] == '%' || s[v[i][1]-(1+x)] == '}'; x++ {
		}
		if v[i][0]+a > len(s) || v[i][1]-x > len(s) {
			break
		}
		e, ok := syscall.Getenv(strings.ToLower(s[v[i][0]+a : v[i][1]-x]))
		if !ok {
			b.WriteString(s[v[i][0]:v[i][1]])
			continue
		}
		b.WriteString(e)
	}
	b.WriteString(s[v[len(v)-1][1]:])
	r := b.Output()
	builders.Put(b)
	return r
}
