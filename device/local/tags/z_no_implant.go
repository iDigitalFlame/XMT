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

package tags

import "github.com/iDigitalFlame/xmt/util"

// Capabilities returns a string version of the 'Enabled' bitmask. Supported
// capabilities will be listed, seperated by semicolons.
func Capabilities() string {
	return ParseCapabilities(setCapWindows, Enabled)
}

// ParseCapabilities returns a string representation of the numerical capabilities
// tag value. The provided boolean will determine if Windows specific capabilities
// are shown. Use false to disable Windows specific capabilities.
func ParseCapabilities(windows bool, v uint32) string {
	var b util.Builder
	if v&CapBugs != 0 {
		b.WriteString("bugtrack;")
	}
	if v&CapImplant != 0 {
		b.WriteString("implant;")
	}
	if v&CapCrypt != 0 {
		b.WriteString("crypt_mapper;")
	}
	if v&CapEws != 0 {
		b.WriteString("encrypt_while_sleep;")
	}
	if v&CapKey != 0 {
		b.WriteString("key_crypt;")
	}
	switch {
	case v&(CapProxy&CapMultiProxy) != 0:
		b.WriteString("proxy_multi;")
	case v&CapProxy != 0:
		b.WriteString("proxy_single;")
	}
	switch {
	case v&CapLimitTiny == CapLimitTiny:
		b.WriteString("limits_tiny;")
	case v&CapLimitStandard == CapLimitStandard:
		b.WriteString("limits_standard;")
	case v&CapLimitLarge == CapLimitLarge:
		b.WriteString("limits_large;")
	case v&CapLimitMedium == CapLimitMedium:
		b.WriteString("limits_medium;")
	default:
		b.WriteString("limits_disabled;")
	}
	if v&CapRandomType == 0 {
		b.WriteString("fast_rand;")
	}
	if v&CapRegexp == 0 {
		b.WriteString("fast_regexp;")
	}
	if v&CapMemorySweeper != 0 {
		b.WriteString("mem_sweep;")
	}
	if windows {
		if v&CapProcEnum != 0 {
			b.WriteString("enum_qsi;")
		} else {
			b.WriteString("enum_snap;")
		}
		if v&CapFuncmap != 0 {
			b.WriteString("funcmap;")
		}
		if v&CapMemoryMapper != 0 {
			b.WriteString("mem_sections;")
		} else {
			b.WriteString("mem_alloc_write;")
		}
		if v&(CapAltLoad|CapCrypt) != 0 {
			b.WriteString("alt_loader;")
		}
		if v&CapChunkHeap != 0 {
			b.WriteString("chunk_heap;")
		}
	}
	if b.Len() == 0 {
		return ""
	}
	s := b.Output()
	return s[:len(s)-1]
}
