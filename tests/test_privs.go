package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/device"
)

func testPrivs() {
	if err := device.AdjustPrivileges("SeShutdownPrivilege", "SeUndockPrivilege"); err != nil {
		panic(err)
	}

	z := cmd.NewProcess("whoami", "/priv")
	o, err := z.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(o))
}
