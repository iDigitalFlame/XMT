package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/iDigitalFlame/xmt/cmd"
	"github.com/iDigitalFlame/xmt/device"
	"github.com/iDigitalFlame/xmt/device/devtools"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "%s <process_name>\n", os.Args[0])
		os.Exit(1)
	}

	devtools.AdjustPrivileges("SeDebugPrivilege")

	device.Local.Hostname = "REDACTED"
	device.Local.User = "LOCALHOST\\REDACTED"
	fmt.Println(device.Local.String())

	c, x := context.WithTimeout(context.Background(), time.Second*10)
	p := cmd.NewProcessContext(c, "cmd", "/c", "echo THIS-IS-A-TEST & ping 127.0.0.1 & echo TEST-DONE & whoami")
	p.SetNoWindow(true)
	p.SetWindowDisplay(0)
	p.SetParentEx(os.Args[1], false)
	o, err := p.CombinedOutput()
	fmt.Printf("result: (err: %s)\n%s\n", err, o)
	x()
}

func main1() {
	fmt.Printf("String: [%s]\nFull: [%s]\nSig: [%s]\nHash: %d, %X",
		device.UUID.String(),
		device.UUID.FullString(),
		device.UUID.Signature(),
		device.UUID.Hash(),
		device.UUID.Hash(),
	)
}
