package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/device"
)

func main3() {
	fmt.Printf(
		"ID:\t\t%s\nOS:\t\t%s\nPID:\t\t%d\nPPID:\t\t%d\nArch:\t\t%s\nUser:\t\t%s\nVersion:\t%s\nHostname:\t%s\nElevated:\t%t\n",
		device.Local.ID, device.Local.OS, device.Local.PID, device.Local.PPID, device.Local.Arch, device.Local.User,
		device.Local.Version, device.Local.Hostname, device.Local.Elevated,
	)

	fmt.Println()

	fmt.Printf(
		"ID:\t\t%s\nOS:\t\t%s\nPID:\t\t%d\nPPID:\t\t%d\nArch:\t\t%s\nUser:\t\t%s\nVersion:\t%s\nHostname:\t%s\nElevated:\t%t\n",
		device.UUID, device.OS, device.PID(), device.PPID(), device.Arch, device.User(), device.Version, device.Hostname(), device.Elevated(),
	)

	r, err := device.Registry("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "SystemRoot")
	fmt.Printf("\nreg: %+v\nerr: %s\n", r, err)

	if err == nil {
		fmt.Printf("Name: %s\nTime: %s\nDir: %t\nLen: %d\n", r.Name(), r.ModTime(), r.IsDir(), r.Size())

		b := make([]byte, r.Size())

		n, err := r.Read(b)
		fmt.Printf("n: %d, err: %s, data[%s]\n", n, err, b)
	}
}
