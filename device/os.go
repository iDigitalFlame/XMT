package device

import (
	"os"
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
	if len(s) >= 2 && s[0] == '~' && s[1] == '/' && len(home) > 0 {
		// Account for shell expansion. (Except JS/WASM)
		s = home + s[1:]
	}
	var (
		l  = -1
		b  = builders.Get().(*util.Builder)
		c  byte
		v  string
		ok bool
	)
	for i := range s {
		switch {
		case s[i] == '$':
			if c > 0 {
				if c == '{' {
					b.WriteString(s[l-1 : i])
				} else {
					b.WriteString(s[l:i])
				}
			}
			c, l = s[i], i
		case s[i] == '%' && c == '%' && i != l:
			if v, ok = syscall.Getenv(s[l+1 : i]); ok {
				b.WriteString(v)
			} else {
				b.WriteString(s[l:i])
			}
			c, l = 0, 0
		case s[i] == '%':
			c, l = s[i], i
		case s[i] == '}' && c == '{':
			if v, ok = syscall.Getenv(s[l+1 : i]); ok {
				b.WriteString(v)
			} else {
				b.WriteString(s[l-1 : i])
			}
			c, l = 0, 0
		case s[i] == '{' && i > 0 && c == '$':
			c, l = s[i], i
		case s[i] >= 'a' && s[i] <= 'z':
			fallthrough
		case s[i] >= 'A' && s[i] <= 'Z':
			fallthrough
		case s[i] == '_':
			if c == 0 {
				b.WriteByte(s[i])
			}
		case s[i] >= '0' && s[i] <= '9':
			if c > 0 && i > l && i-l == 1 {
				c, l = 0, 0
			}
			if c == 0 {
				b.WriteByte(s[i])
			}
		default:
			if c == '$' {
				if v, ok = syscall.Getenv(s[l+1 : i]); ok {
					b.WriteString(v)
				} else {
					b.WriteString(s[l:i])
				}
				c, l = 0, 0
			} else if c > 0 {
				if c == '{' {
					b.WriteString(s[l-1 : i])
				} else {
					b.WriteString(s[l:i])
				}
				c, l = 0, 0
			}
			b.WriteByte(s[i])
		}
	}
	if l == -1 {
		builders.Put(b)
		return s
	}
	if l < len(s) && c > 0 {
		if c == '$' {
			if v, ok = syscall.Getenv(s[l+1:]); ok {
				b.WriteString(v)
			} else {
				b.WriteString(s[l:])
			}
		} else if c == '{' {
			b.WriteString(s[l-1:])
		} else {
			b.WriteString(s[l:])
		}
	}
	v = b.Output()
	builders.Put(b)
	return v
}
