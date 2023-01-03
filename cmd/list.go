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

import "github.com/iDigitalFlame/xmt/data"

// ProcessInfo is a struct that holds simple process related data for extraction
// This struct is returned via a call to 'Processes'.
//
// This struct also supports binary Marshaling/UnMarshaling.
type ProcessInfo struct {
	_          [0]func()
	Name, User string
	PID, PPID  uint32
}
type processList []ProcessInfo

func (p processList) Len() int {
	return len(p)
}
func (p processList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p processList) Less(i, j int) bool {
	if p[i].PPID == p[j].PPID {
		return p[i].PID < p[j].PID
	}
	return p[i].PPID < p[j].PPID
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (p ProcessInfo) MarshalStream(w data.Writer) error {
	if err := w.WriteUint32(p.PID); err != nil {
		return err
	}
	if err := w.WriteUint32(p.PPID); err != nil {
		return err
	}
	if err := w.WriteString(p.Name); err != nil {
		return err
	}
	return w.WriteString(p.User)
}

// UnmarshalStream transforms this struct from a binary format that is read from
// the supplied data.Reader.
func (p *ProcessInfo) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint32(&p.PID); err != nil {
		return err
	}
	if err := r.ReadUint32(&p.PPID); err != nil {
		return err
	}
	if err := r.ReadString(&p.Name); err != nil {
		return err
	}
	return r.ReadString(&p.User)
}
