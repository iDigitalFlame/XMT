//go:build windows

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

package cmd

import (
	"sort"

	"github.com/iDigitalFlame/xmt/device/winapi"
	"github.com/iDigitalFlame/xmt/util/bugtrack"
)

// Processes attempts to gather the current running Processes and returns them
// as a slice of ProcessInfo structs, otherwise any errors are returned.
func Processes() ([]ProcessInfo, error) {
	if err := winapi.GetDebugPrivilege(); err != nil {
		if bugtrack.Enabled {
			bugtrack.Track("cmd.Processes(): GetDebugPrivilege failed with err=%s", err.Error())
		}
	}
	var (
		r   = make(processList, 0, 64)
		err = winapi.EnumProcesses(func(e winapi.ProcessEntry) error {
			u, _ := e.User()
			r = append(r, ProcessInfo{Name: e.Name, User: u, PID: e.PID, PPID: e.PPID})
			return nil
		})
	)
	sort.Sort(r)
	return r, err
}
