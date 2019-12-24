package com

import (
	"errors"
	"net"
)

// ErrInvalidNetwork is an error returned from the NewStreamConnector function
// when a non-stream network is used, or the NewChunkConnector function when a stream
// network is used.
var ErrInvalidNetwork = errors.New("invalid network type")

// Server is an interface that is used to Listen on a specific protocol
// for client connections.  The Listener does not take any actions on the clients
// but transcribes the data into bytes for the Session handler.
type Server interface {
	Listen(string) (net.Listener, error)
}

// Provider is a combined interface that inherits the Server and Connector
// interfaces. This allows a single struct to provide listening and connection
// capabilities.
type Provider interface {
	Connector
	Server
}

// Connector is an interface that passes methods that can be used to form
// connections between the client and server.  Other functions include the
// process of listening and accepting connections.
type Connector interface {
	Connect(string) (net.Conn, error)
}
