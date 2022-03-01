package main

import (
	"os"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/cmd/filter"
)

func testElevated() {

	if len(os.Args) <= 1 {
		os.Stderr.WriteString("usage: " + os.Args[0] + " <command>\n")
		os.Exit(2)
	}

	switch os.Args[1] {
	case "/?", "?", "-h", "-?":
		os.Stderr.WriteString("usage: " + os.Args[0] + " <command>\n")
		os.Exit(2)
	default:
	}

	x := cmd.NewProcess(os.Args[1:]...)
	x.SetParent(&filter.Filter{Include: []string{"TrustedInstaller.exe"}, Elevated: filter.True})

	b, err := x.CombinedOutput()

	if err != nil {
		if _, ok := err.(*cmd.ExitError); !ok {
			panic(err)
		}
	}

	os.Stdout.Write(b)
	os.Stdout.WriteString("\n")

}
