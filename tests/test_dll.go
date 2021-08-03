package main

import (
	"os"
	"path/filepath"

	"github.com/iDigitalFlame/xmt/cmd"
)

func TestDLL() {
	var (
		e, _ = os.Executable()
		p    = filepath.Dir(e)
		d    = filepath.Join(p, os.Args[1])
	)

	os.Stdout.WriteString("DLL at: " + d + "\n")

	c := cmd.NewDLL(d)
	if len(os.Args) >= 3 {
		c.SetParent(cmd.F().SetInclude(os.Args[2]))
	} else {
		c.SetParent(cmd.RandomParent)
	}

	var (
		err   = c.Run()
		_, ok = err.(*cmd.ExitError)
	)

	if c.Stop(); !ok && err != nil {
		panic(err)
	}
}
