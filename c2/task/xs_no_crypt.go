//go:build scripts && !crypt

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

import "github.com/iDigitalFlame/xmt/device/local"

func createEnvironment() map[string]any {
	return map[string]any{
		"ID":       local.UUID.String(),
		"OS":       local.Version,
		"PID":      local.Device.PID,
		"PPID":     local.Device.PPID,
		"ADMIN":    local.Elevated(),
		"HOSTNAME": local.Device.Hostname,
	}
}
