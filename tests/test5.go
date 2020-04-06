package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/c2"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/data/crypto"
)

func main() {
	c := c2.Config{
		c2.Sleep(time.Second * 5),
		c2.Jitter(30),
		c2.ConnectTCP,
		c2.Size(256),
		c2.WrapGzipLevel(2),
		c2.TransformBase64,
	}

	fmt.Printf("%s\n", c)

	for i := range c {
		fmt.Printf("\t%s\n", c[i])
	}

}

func main11() {

	a, b := crypto.NewCBK(10), crypto.NewCBK(10)

	//a.Source, b.Source = crypto.NewSource("this is my password"), crypto.NewSource("this is my password")

	a.A, b.A = 30, 30
	a.B, b.B = 6, 6
	a.C, b.C = 90, 90

	c := new(data.Chunk)
	c.Limit = 64

	n, err := a.Write(c, []byte("This is my super secure password! 1234567890"))
	fmt.Printf("n1, err: %d, %s\n", n, err)

	a.Flush(c)

	fmt.Printf("out: %d [%s]\n", c.Len(), c.Payload())

	o, r := make([]byte, n+10), bytes.NewReader(c.Payload())

	n, err = b.Read(r, o)
	fmt.Printf("n, err: %d, %s\n", n, err)

	fmt.Printf("res: %s\n", o)

}
