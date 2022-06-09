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

// Package xerr is a simplistic (and more efficient) re-write of the "errors"
// built-in package.
//
// This is used to create comparable and (sometimes) un-wrapable error structs.
//
// This package acts differently when the "implant" build tag is used. If enabled,
// Most error string values are stripped to prevent identification and debugging.
//
// It is recommended if errors are needed to be compared even when in an implant
// build, to use the "Sub" function, which will ignore error strings and use
// error codes instead.
//
package xerr
