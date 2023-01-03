// Copyright (C) 2020 - 2023 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package winapi

import "testing"

func TestFnv(t *testing.T) {
	v := [...]struct {
		Value string
		Hash  uint32
	}{
		{"NtTraceEvent", 0x89F984CE},
		{"NtCancelIoFileEx", 0xD4909C18},
		{"RtlCopyMappedMemory", 0x381752E6},
		{"NtUnmapViewOfSection", 0x19B022D},
		{"NtWriteVirtualMemory", 0x2012F428},
	}
	for i := range v {
		if r := FnvHash(v[i].Value); r != v[i].Hash {
			t.Fatalf(`TestFnv(): FnvHash result for "%s" "0x%X" did not match "0x%X" !`, v[i].Value, r, v[i].Hash)
		}
	}
}
func TestStrings(t *testing.T) {
	v := [...]string{
		"hello test1",
		"test1234",
		"example12345",
		"string value123",
	}
	for i := range v {
		r, err := UTF16FromString(v[i])
		if err != nil {
			t.Fatalf(`TestStrings(): UTF16FromString failed for string "%s": %s!`, v[i], err.Error())
		}
		if len(r) != len(v[i])+1 {
			t.Fatalf(`TestStrings(): UTF16FromString result for string "%s" / "%d" was not the expected size "%d"!`, v[i], len(r), len(v[i])+1)
		}
		if k := UTF16ToString(r); k != v[i] {
			t.Fatalf(`TestStrings(): UTF16ToString result "%s" does not match "%s"!`, k, v[i])
		}
		p, err := UTF16PtrFromString(v[i])
		if err != nil {
			t.Fatalf(`TestStrings(): UTF16PtrFromString failed for string "%s": %s!`, v[i], err.Error())
		}
		if k := UTF16PtrToString(p); k != v[i] {
			t.Fatalf(`TestStrings(): UTF16PtrToString result "%s" does not match "%s"!`, k, v[i])
		}
	}
}
