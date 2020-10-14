package main

import (
	"github.com/iDigitalFlame/xmt/device"
)

func main() {
	device.Expand("testing SHELL=${shell}bin; $home lol${pwd}dd")
}
