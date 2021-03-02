package device

import (
	"os"
	"regexp"
	"runtime"
	"strings"
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
	// ArchUnknown represents an unknown chipset family.
	ArchUnknown deviceArch = 0x5
)

var envExp = regexp.MustCompile(`%[\w\d()_-]+%|\$[\w\d_-]+|\$\{[[\w\d_-]+\}`)

type deviceOS uint8
type deviceArch uint8

func getArch() deviceArch {
	switch runtime.GOARCH {
	case "386":
		return Arch86
	case "amd64", "amd64p32":
		return Arch64
	case "ppc", "ppc64", "ppc64le":
		return ArchPowerPC
	case "arm", "armbe", "arm64", "arm64be":
		return ArchARM
	case "mips", "mipsle", "mips64", "mips64le", "mips64p32", "mips64p32le":
		return ArchMips
	}
	return ArchUnknown
}

// Expand attempts to determine environment variables from the current session and translate them from
// the supplied string. This function supports both Windows (%var%) and *nix ($var or ${var}) variable substitutions.
func Expand(s string) string {
	v := envExp.FindAllStringIndex(s, -1)
	if len(v) == 0 {
		return s
	}
	b := builders.Get().(*strings.Builder)
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
		e, ok := Environment[strings.ToLower(s[v[i][0]+a:v[i][1]-x])]
		if !ok {
			b.WriteString(s[v[i][0]:v[i][1]])
			continue
		}
		b.WriteString(e)
	}
	b.WriteString(s[v[len(v)-1][1]:])
	r := b.String()
	b.Reset()
	builders.Put(b)
	return r
}
func getEnv() map[string]string {
	m := make(map[string]string)
	for _, v := range os.Environ() {
		if i := strings.IndexByte(v, 61); i > 0 {
			m[strings.ToLower(v[:i])] = v[i+1:]
		}
	}
	t := os.TempDir()
	m["tmp"], m["temp"], m["tmpdir"], m["tempdir"] = t, t, t, t
	return m
}
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
	case ArchMips:
		return "MIPS"
	case ArchPowerPC:
		return "PowerPC"
	}
	return "Unknown"
}
