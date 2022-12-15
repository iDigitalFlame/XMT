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

// Package local contains many functions and variables that contain information
// about the local device.
package local

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"os"

	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/local/tags"
)

var (
	// UUID is the device specific and session specific identifier.
	UUID = getID()

	// Version is the local machine's Operating System version information.
	Version = version()
)

// Device is the pointer to the local machine instance. This instance is loaded
// at runtime and is used for local data gathering and identification.
var Device = (&local{Machine: device.Machine{
	ID:           UUID,
	PID:          uint32(os.Getpid()),
	PPID:         getPPID(),
	User:         getUsername(),
	System:       systemType(),
	Version:      Version,
	Network:      make(device.Network, 0),
	Elevated:     isElevated(),
	Capabilities: tags.Enabled,
}}).init()

type local struct {
	/* * */ device.Machine
}

// Elevated will return true if the current process has elevated privileges,
// false otherwise.
//
// This function is evaluated at runtime.
func Elevated() bool {
	return Device.Elevated&1 == 1
}
func getID() device.ID {
	var i device.ID
	if s := sysID(); len(s) > 0 {
		h := hmac.New(sha256.New, s)
		h.Write([]byte(vers))
		copy(i[:], h.Sum(nil))
		h.Reset()
	} else {
		rand.Read(i[:])
	}
	// NOTE(dij): This code below changes how IDs are generated
	//            An extra bit is taken away from the random address space
	//            (8 => 7), thus short IDs from the same machine will ALWAYS
	//            have the same two first bits for easy identification.
	if rand.Read(i[device.MachineIDSize+1:]); i[0] == 0 {
		i[0] = 1
	}
	if i[device.MachineIDSize] == 0 {
		i[device.MachineIDSize] = 1
	}
	return i
}
func (l *local) init() *local {
	l.Hostname, _ = os.Hostname()
	l.Network.Refresh()
	l.fixHostname()
	return l
}
func (l *local) fixHostname() {
	for c, i := 0, 0; i < len(l.Hostname); i++ {
		if l.Hostname[i] != '.' {
			continue
		}
		if c++; c > 1 {
			l.Hostname = l.Hostname[:i]
			return
		}
	}
}
func (l *local) Refresh() error {
	l.User = getUsername()
	var err error
	if l.Hostname, err = os.Hostname(); err != nil {
		return err
	}
	if err := l.Network.Refresh(); err != nil {
		return err
	}
	l.PID = uint32(os.Getpid())
	l.PPID = getPPID()
	l.Elevated = isElevated()
	l.fixHostname()
	return nil
}
