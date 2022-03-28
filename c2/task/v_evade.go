//go:build !implant

package task

import "github.com/iDigitalFlame/xmt/com"

// CheckDLL is a similar function to ReloadDLL.
// This function will return 'true' if the contents in memory match the
// contents of the file on disk. Otherwise it returns false.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvCheckDLL
//
//  Input:
//      string // DLL Name
//  Output:
//      bool   // DLL Check result, true if normal.
//
// C2 Client Command:
//  dll_check <name>
func CheckDLL(d string) *com.Packet {
	n := &com.Packet{ID: TvCheckDLL}
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
//  ID: TvReloadDLL
//
//  Input:
//      string // DLL Name
//  Output:
//      <none>
//
// C2 Client Command:
//  dll_reload <name>
func ReloadDLL(d string) *com.Packet {
	n := &com.Packet{ID: TvReloadDLL}
	n.WriteString(d)
	return n
}
