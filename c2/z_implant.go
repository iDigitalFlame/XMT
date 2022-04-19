//go:build implant

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
type proxyState struct{}

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
func (*Server) Remove(_ device.ID, _ bool)     {}
func (*Session) updateProxyInfo(_ []proxyData) {}
