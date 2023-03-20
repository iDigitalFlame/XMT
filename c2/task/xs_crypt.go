//go:build scripts && crypt && go1.18
// +build scripts,crypt,go1.18

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

package task

import (
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func createEnvironment() map[string]any {
	return map[string]any{
		crypt.Get(1): local.UUID.String(),   // ID
		crypt.Get(2): local.Version,         // OS
		crypt.Get(3): local.Device.PID,      // PID
		crypt.Get(4): local.Device.PPID,     // PPID
		crypt.Get(5): local.Elevated(),      // ADMIN
		crypt.Get(6): local.Device.Hostname, // HOSTNAME
	}
}
