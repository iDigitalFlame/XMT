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

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
)

func testElevated() {

	if len(os.Args) <= 1 {
		os.Stderr.WriteString("usage: " + os.Args[0] + " <command>\n")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "/?", "?", "-h", "-?":
		os.Stderr.WriteString("usage: " + os.Args[0] + " <command>\n")
		os.Exit(2)
	default:
	}

	x := cmd.NewProcess(os.Args[1:]...)
	x.SetParent(&filter.Filter{Include: []string{"TrustedInstaller.exe"}, Elevated: filter.True})

	b, err := x.CombinedOutput()

	if err != nil {
		if _, ok := err.(*cmd.ExitError); !ok {
			panic(err)
		}
	}

	os.Stdout.Write(b)
	os.Stdout.WriteString("\n")

}
