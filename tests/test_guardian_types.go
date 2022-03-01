package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/man"
)

func testGuardianTypes() {

	if len(os.Args) == 1 {

		fmt.Printf("g1 = %t\n", man.Check(man.Pipe, "1guard1"))
		fmt.Printf("g2 = %t\n", man.Check(man.TCP, ":5011"))
		fmt.Printf("g3 = %t\n", man.Check(man.Mutex, "1guard2"))
		fmt.Printf("g4 = %t\n", man.Check(man.Event, "1guard3"))
		fmt.Printf("g5 = %t\n", man.Check(man.Semaphore, "1guard4"))
		fmt.Printf("g6 = %t\n", man.Check(man.Mailslot, "1guard5"))

		man.Sentinel{
			Linker: "mutex",
			Filter: filter.F().SetInclude("derp.exe", "explorer.exe").SetElevated(true),
			Paths:  []string{"cmd1.exe", "http://google.com", "cc.exe"},
		}.File("a.txt")

		s := man.F("a.txt")
		fmt.Println(s, s.Filter.Include)
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
