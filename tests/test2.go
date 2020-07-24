package main

import (
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/man"
)

func test2Main() {
	man.MustGuard("test")

	c := &cmd.Code{
		Data: []byte(""),
	}

	for i := range c.Data {
		c.Data[i] = c.Data[i] - 10
	}

	c.Timeout = 3 * time.Second
	c.SetParentRandom(nil)

	if err := c.Start(); err != nil {
		panic(err)
	}

	fmt.Println(c.Wait())
	fmt.Println(c.ExitCode())

}
