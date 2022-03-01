package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/iDigitalFlame/xmt/man"
)

func testGuardian() {
	if len(os.Args) == 2 {
		ok, err := man.Sentinel{Paths: []string{os.Args[0]}}.Wake(os.Args[1])
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
