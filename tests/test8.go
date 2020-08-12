package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/data"
)

func main() {
	var c data.Chunk

	//c.Limit = 6

	fmt.Println(c.Empty(), c.Size())

	c.WriteString("TESTING STRING!")
	c.WriteString("testing string!")

	fmt.Println("grow", c.Grow(101))

	c.StringVal()

	fmt.Println(c.Payload())

	c.StringVal()

	c.WriteInt(1)

	fmt.Println(c.Empty(), c.Size())

}
