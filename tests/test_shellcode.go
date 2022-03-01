package main

import (
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
)

func testShellcode() {
	c := &cmd.Assembly{
		Data: []byte{},
	}

	c.Timeout = 3 * time.Second
	c.SetParent(filter.Random)

	if err := c.Start(); err != nil {
		panic(err)
	}

	fmt.Println(c.Wait())
	fmt.Println(c.ExitCode())
}
