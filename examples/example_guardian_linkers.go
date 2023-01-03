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

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iDigitalFlame/xmt/man"
)

func exampleGuardianLinkers() {
	if len(os.Args) == 1 {
		fmt.Printf("g1 = %t\n", man.Check(man.Pipe, "1guard1"))
		fmt.Printf("g2 = %t\n", man.Check(man.TCP, ":5011"))
		fmt.Printf("g3 = %t\n", man.Check(man.Mutex, "1guard2"))
		fmt.Printf("g4 = %t\n", man.Check(man.Event, "1guard3"))
		fmt.Printf("g5 = %t\n", man.Check(man.Semaphore, "1guard4"))
		fmt.Printf("g6 = %t\n", man.Check(man.Mailslot, "1guard5"))

		var s man.Sentinel
		s.SetInclude("derp.exe", "explorer.exe").SetElevated(true)
		s.AddExecute("cmd1.exe")
		s.AddDownload("http://google.com")
		s.AddExecute("cc.exe")
		s.Save(nil, "a.txt")

		v, _ := man.File(nil, "a.txt")
		fmt.Println(v, v.Include)
		return
	}

	g1 := man.MustGuard(man.Pipe, "1guard1")
	g2 := man.MustGuard(man.TCP, ":5011")
	g3 := man.MustGuard(man.Mutex, "1guard2")
	g4 := man.MustGuard(man.Event, "1guard3")
	g5 := man.MustGuard(man.Semaphore, "1guard4")
	g6 := man.MustGuard(man.Mailslot, "1guard5")

	e := make(chan os.Signal, 1)
	signal.Notify(e, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	<-e

	g1.Close()
	g2.Close()
	g3.Close()
	g4.Close()
	g5.Close()
	g6.Close()
}
