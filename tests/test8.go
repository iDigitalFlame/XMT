package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/util/xerr"
)

func main() {
	var (
		d  = xerr.New("test error")
		d1 = xerr.New("test error")
		d2 = xerr.New("test error2")
	)
	fmt.Println(d, d1, d2, "\n", d == d, d == d1, d == d2)
}
