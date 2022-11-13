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

// Package tags enables identification of the build tags and capabilities that
// are compiled into the current program. This package provides only top-level
// functions that are read-only.
package tags

// Cap* capability flag constants.
const (
	CapMemoryMapper uint32 = 1 << iota
	CapBugs
	CapImplant
	CapCrypt
	CapEws
	CapKey
	CapProxy
	CapMultiProxy // += CapProxy
	CapProcEnum
	CapRegexp
	CapLimitLarge
	CapLimitMedium
	CapLimitSmall
	CapRandomType
	CapMemorySweeper
	CapFuncmap
	CapAltLoad

	// CapLimitStandard is a flagset combo that is recognized as
	//  CapLimitLarge | CapLimitMedium
	// and should be tested before CapLimitLarge and CapLimitMedium or tested
	// against it's value.
	//  standard := tags.Enabled & CapLimitStandard == CapLimitStandard
	CapLimitStandard = CapLimitLarge | CapLimitMedium
	// CapLimitTiny is a flagset combo that is recognized as
	//   CapLimitLarge | CapLimitMedium | CapLimitSmall
	// and should be tested before CapLimitLarge, CapLimitMedium and CapLimitSmall
	// or tested against it's value.
	//  tiny := tags.Enabled & CapLimitTiny == CapLimitTiny
	CapLimitTiny = CapLimitLarge | CapLimitMedium | CapLimitSmall
)

// Enabled contains a bitmask combination of all the enabled Capabilities (build
// tags) enabled currently.
//
// Windows specific flags may exist but only take effect if the device is running
// Windows.
const Enabled = setCapAltLoad | setCapFuncmap | setCapMemorySweeper |
	setCapRandomType | setCapLimit | setCapRegexp | setCapProcEnum |
	setCapProxy | setCapKey | setCapEws | setCapCrypt | setCapImplant |
	setCapBugs | setCapMemoryMapper
