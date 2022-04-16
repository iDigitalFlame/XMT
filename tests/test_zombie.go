package main

import (
	"fmt"
	"os"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
)

func testZombie() {

	b, err := os.ReadFile(`my.dll`)
	if err != nil {
		panic(err)
	}

	z := cmd.NewZombie(cmd.DLLToASM("", b), "notepad.exe", "file.txt")
	z.SetParent(filter.I("sihost.exe"))

	if err = z.Start(); err != nil {
		panic(err)
	}

	if err = z.Wait(); err != nil {
		panic(err)
	}

	e, err := z.ExitCode()
	if err != nil {
		panic(err)
	}

	fmt.Printf("res: %X\n", uint32(e))
}
