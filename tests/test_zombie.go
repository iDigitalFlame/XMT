package main

import (
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
)

func funcZombie() {

	z := cmd.NewZombie([]byte{}, "notepad.exe", "this-is-not-real.txt")
	z.SetParent(filter.I("sihost.exe"))
	z.SetSuspended(true)

	if err := z.Start(); err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)
	println("resume NOW!")

	z.Resume()
	if err := z.Wait(); err != nil {
		panic(err)
	}
	e, err := z.ExitCode()
	if err != nil {
		panic(err)
	}

	fmt.Printf("res: %d\n", e)
}
