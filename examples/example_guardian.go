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
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/iDigitalFlame/xmt/man"
)

func exampleGuardian() {
	if len(os.Args) == 2 {
		var s man.Sentinel
		s.AddExecute(man.Self)
		ok, err := s.Wake(man.Pipe, os.Args[1])
		if err != nil {
			fmt.Println("error", err)
		}
		if ok {
			fmt.Println("launched!")
		}

		fmt.Println("pinged!")
		return
	}

	var (
		x, c   = context.WithCancel(context.Background())
		g, err = man.GuardContext(x, man.Pipe, "testguard")
	)
	if err != nil {
		panic(err)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s)

	<-s
	c()
	g.Wait()
}
