// Copyright (C) 2020 - 2022 iDigitalFlame
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

import (
	"syscall"
	"unsafe"
)

const (
	utfSelf        = 0x10000
	utfSurgA       = 0xd800
	utfSurgB       = 0xdc00
	utfSurgC       = 0xe000
	utfRuneMax     = '\U0010FFFF'
	utfReplacement = '\uFFFD'
)

// SliceHeader is the runtime representation of a slice.
//
// It cannot be used safely or portably and its representation may change in a
// later release.
//
// ^ Hey, shut up.
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// UTF16Decode returns the Unicode code point sequence represented by the UTF-16
// encoding rune values supplied.
func UTF16Decode(s []uint16) []rune {
	var (
		b = make([]rune, len(s))
		n int
	)
loop:
	for i := 0; i < len(s); i++ {
		switch r := s[i]; {
		case r == 0:
			break loop
		case r < utfSurgA, utfSurgC <= r:
			b[n] = rune(r)
		case utfSurgA <= r && r < utfSurgB && i+1 < len(s) && utfSurgB <= s[i+1] && s[i+1] < utfSurgC:
			b[n] = utf16DecodeRune(rune(r), rune(s[i+1]))
			i++
		default:
			b[n] = utfReplacement
		}
		n++
	}
	return b[:n]
}

// UTF16ToString returns the UTF-8 encoding of the UTF-16 sequence s, with a
// terminating NUL and any bytes after the NUL removed.
func UTF16ToString(s []uint16) string {
	return string(UTF16Decode(s))
}
func utf16DecodeRune(r1, r2 rune) rune {
	if utfSurgA <= r1 && r1 < utfSurgB && utfSurgB <= r2 && r2 < utfSurgC {
		return (r1-utfSurgA)<<10 | (r2 - utfSurgB) + utfSelf
	}
	return utfReplacement
}

// UTF16EncodeStd encodes the runes into a UTF16 array and ignores zero points.
//
// This is ONLY safe to use if you know what you're doing.
func UTF16EncodeStd(s []rune) []uint16 {
	n := len(s)
	for i := range s {
		if s[i] < utfSelf {
			continue
		}
		n++
	}
	var (
		b = make([]uint16, n)
		i int
	)
	for n = 0; n < len(s); i++ {
		switch {
		case 0 <= s[i] && s[i] < utfSurgA, utfSurgC <= s[i] && s[i] < utfSelf:
			b[n] = uint16(s[i])
			n++
		case utfSelf <= s[i] && s[i] <= utfRuneMax:
			b[n], b[n+1] = utf16EncodeRune(s[i])
			n += 2
		default:
			b[n] = uint16(utfReplacement)
			n++
		}
	}
	return b[:n]
}

// UTF16PtrToString takes a pointer to a UTF-16 sequence and returns the
// corresponding UTF-8 encoded string.
//
// If the pointer is nil, it returns the empty string. It assumes that the UTF-16
// sequence is terminated at a zero word; if the zero word is not present, the
// program may crash.
func UTF16PtrToString(p *uint16) string {
	if p == nil || *p == 0 {
		return ""
	}
	n := 0
	for ptr := unsafe.Pointer(p); *(*uint16)(ptr) != 0; n++ {
		ptr = unsafe.Pointer(uintptr(ptr) + unsafe.Sizeof(*p))
	}
	var s []uint16
	h := (*SliceHeader)(unsafe.Pointer(&s))
	h.Data, h.Len, h.Cap = unsafe.Pointer(p), n, n
	return string(UTF16Decode(s))
}
func utf16Encode(s []rune) ([]uint16, error) {
	n := len(s)
	for i := range s {
		if s[i] == 0 && i+1 < len(s) {
			return nil, syscall.EINVAL
		}
		if s[i] < utfSelf {
			continue
		}
		n++
	}
	var (
		b = make([]uint16, n)
		i int
	)
	for n = 0; n < len(s); i++ {
		switch {
		case s[i] == 0 && i+1 < len(s):
			return nil, syscall.EINVAL
		case 0 <= s[i] && s[i] < utfSurgA, utfSurgC <= s[i] && s[i] < utfSelf:
			b[n] = uint16(s[i])
			n++
		case utfSelf <= s[i] && s[i] <= utfRuneMax:
			b[n], b[n+1] = utf16EncodeRune(s[i])
			n += 2
		default:
			b[n] = uint16(utfReplacement)
			n++
		}
	}
	return b[:n], nil
}
func utf16EncodeRune(r rune) (uint16, uint16) {
	if r < utfSelf || r > utfRuneMax {
		return utfReplacement, utfReplacement
	}
	r -= utfSelf
	return uint16(utfSurgA + (r>>10)&0x3FF), uint16(utfSurgB + r&0x3FF)
}

// UTF16FromString returns the UTF-16 encoding of the UTF-8 string with a
// terminating NUL added.
//
// If the string contains a NUL byte at any location, it returns syscall.EINVAL.
func UTF16FromString(s string) ([]uint16, error) {
	if len(s) == 0 {
		return []uint16{0}, nil
	}
	return utf16Encode([]rune(s + "\x00"))
}

// UTF16PtrFromString returns pointer to the UTF-16 encoding of the UTF-8 string,
// with a terminating NUL added.
//
// If the string contains a NUL byte at any location, it returns syscall.EINVAL.
func UTF16PtrFromString(s string) (*uint16, error) {
	a, err := UTF16FromString(s)
	if err != nil {
		return nil, err
	}
	return &a[0], nil
}
