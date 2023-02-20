//go:build implant
// +build implant

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

package c2

import (
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

const maxEvents = 512

// Server is the manager for all C2 Listener and Sessions connection and states.
// This struct also manages all events and connection changes.
type Server struct{}

// Listener is a struct that is passed back when a C2 Listener is added to the
// Server.
//
// The Listener struct allows for controlling the Listener and setting callback
// functions to be used when a client connects, registers or disconnects.
type Listener struct{}

func (*Listener) oneshot(_ *com.Packet) {}

// Remove removes and closes the Session and releases all it's associated
// resources from this server instance.
//
// If shutdown is false, this does not close the Session on the client's end and
// will just remove the entry, but can be re-added and if the client connects
// again.
//
// If shutdown is true, this will trigger a Shutdown packet to be sent to close
// down the client and will wait until the client acknowledges the shutdown
// request before removing.
func (*Server) Remove(_ device.ID, _ bool) {}
