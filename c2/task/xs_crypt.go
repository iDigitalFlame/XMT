//go:build scripts && crypt

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

package task

import (
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func createEnvironment() map[string]any {
	return map[string]any{
		crypt.Get(8):  local.UUID.String(),   // ID
		crypt.Get(9):  local.Version,         // OS
		crypt.Get(10): local.Device.PID,      // PID
		crypt.Get(11): local.Device.PPID,     // PPID
		crypt.Get(12): local.Elevated(),      // ADMIN
		crypt.Get(13): local.Device.Hostname, // HOSTNAME
	}
}
