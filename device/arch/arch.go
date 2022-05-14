package arch

const (
	// X64 represents the 64-bit chipset family.
	X64 Architecture = 0x0
	// X86 represents the 32-bit chipset family.
	X86 Architecture = 0x1
	// ARM represents the ARM chipset family.
	ARM Architecture = 0x2
	// PowerPC represents the PowerPC chipset family.
	PowerPC Architecture = 0x3
	// Mips represents the MIPS chipset family.
	Mips Architecture = 0x4
	// Risc represents the RiscV chipset family.
	Risc Architecture = 0x5
	// ARM64 represents the ARM64 chipset family.
	ARM64 Architecture = 0x6
	// WASM represents the WASM/JavaScript software family.
	WASM Architecture = 0x7
	// Unknown represents an unknown chipset family.
	Unknown Architecture = 0x8
)

// Architecture is a number representation of the chipset architecture.
type Architecture uint8
