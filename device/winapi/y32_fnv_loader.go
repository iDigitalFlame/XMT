//go:build windows && (altload || crypt) && (arm || 386)

package winapi

type imageOptionalHeader struct {
	Magic               uint16
	_, _                uint8
	SizeOfCode          uint32
	_, _                uint32
	AddressOfEntryPoint uint32
	BaseOfCode          uint32
	_                   uint32
	ImageBase           uint32
	_, _                uint32
	_, _, _, _, _, _    uint16
	_                   uint32
	SizeOfImage         uint32
	SizeOfHeaders       uint32
	_                   uint32
	Subsystem           uint16
	DllCharacteristics  uint16
	_, _, _, _          uint32
	LoaderFlags         uint32
	_                   uint32
	Directory           [16]imageDataDirectory
}
