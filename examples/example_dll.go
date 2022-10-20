//go:build windows

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

package main

import (
	"os"
	"path/filepath"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
)

func exampleDLL() {
	var (
		e, _ = os.Executable()
		p    = filepath.Dir(e)
		d    = filepath.Join(p, os.Args[1])
	)

	os.Stdout.WriteString("DLL at: " + d + "\n")

	c := cmd.NewDLL(d)
	if len(os.Args) >= 3 {
		c.SetParent(filter.F().SetInclude(os.Args[2]))
	} else {
		c.SetParent(filter.Random)
	}

	var (
		err   = c.Run()
		_, ok = err.(*cmd.ExitError)
	)

	if c.Stop(); !ok && err != nil {
		panic(err)
	}
}
