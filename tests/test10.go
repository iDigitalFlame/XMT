package main

import (
	"os"
	"path/filepath"

	"github.com/iDigitalFlame/xmt/cmd"
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
	/*fmt.Println(smonkey.Invoke(script))

	fmt.Println(device.Local.String())

	for _, v := range device.Local.Network {
		fmt.Println(v, v.Mac, v.Address[0], v.Address[0].IP().String())
	}*/

	var (
		e, _ = os.Executable()
		p    = filepath.Dir(e)
		d    = filepath.Join(p, os.Args[1])
	)

	os.Stdout.WriteString("DLL at: " + d + "\n")

	c := cmd.NewDll(d)
	if len(os.Args) >= 3 {
		c.SetParent(os.Args[2])
	}

	var (
		err   = c.Run()
		_, ok = err.(*cmd.ExitError)
	)

	if c.Stop(); !ok && err != nil {
		panic(err)
	}

}
