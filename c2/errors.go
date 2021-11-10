package c2

import "github.com/iDigitalFlame/xmt/util/xerr"

var (
	// ErrNoHost is a error returned by the Connect and Listen functions when the host is empty and the
	// provided Profile is also nil or does not contain a Host hint.
	ErrNoHost = xerr.New("invalid or missing connector")
	// ErrNoConnector is a error returned by the Connect and Listen functions when the Connector is nil and the
	// provided Profile is also nil or does not contain a connection hint.
	ErrNoConnector = xerr.New("invalid or missing connector")

	// ErrMalformedPacket is an error returned by various Packet reading functions when a Packet is
	// attempted to be passed that is invalid. Invalid Packets are packets that do not have a proper ID value
	// or contain an empty device ID.
	ErrMalformedPacket = xerr.New("malformed or invalid Packet")

	// ErrUnable is an error returned for a generic action if there is some condition that prevents the action
	// from running.
	ErrUnable = xerr.New("cannot preform this action")
)
