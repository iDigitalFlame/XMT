package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/cmd/script/smonkey"
	"github.com/iDigitalFlame/xmt/device"
)

const script = `println("Trying to sleep for 2 seconds!")
sleep(2)
println("Done!")
print("Trying to run \"ls -al\" in the current DIR = " + exec("pwd") + "!\n")
val = exec("ls -al")
printf("output of ls is = [\n%s\n]\n", val)
println("Done!")
`

func main() {
	fmt.Println(smonkey.Invoke(script))

	fmt.Println(device.Local.String())

	for _, v := range device.Local.Network {
		fmt.Println(v, v.Mac, v.Address[0], v.Address[0].IP().String())
	}

}
