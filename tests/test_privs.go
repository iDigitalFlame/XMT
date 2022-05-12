package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/device/winapi"
)

func testPrivs() {
	if err := winapi.EnablePrivileges("SeShutdownPrivilege", "SeUndockPrivilege"); err != nil {
		panic(err)
	}

	z := cmd.NewProcess("whoami", "/priv")
	o, err := z.CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(o))
}
