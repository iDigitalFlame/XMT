package main

import (
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
)

func testShellcode() {
	c := &cmd.Code{
		Data: []byte{},
	}

	c.Timeout = 3 * time.Second
	c.SetParent(cmd.RandomParent)

	if err := c.Start(); err != nil {
		panic(err)
	}

	fmt.Println(c.Wait())
	fmt.Println(c.ExitCode())
}
