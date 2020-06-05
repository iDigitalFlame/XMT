package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/com/pipe"
	"github.com/iDigitalFlame/xmt/data/crypto"
)

func test6Main() {
	l, err := pipe.ListenPerms("/tmp/derp.l", "uxrrwx")
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 5)
	l.Close()
}

func test6Main2() {
	var (
		x = crypto.XOR("derpmaster123456")
		v = []byte("this is my key#!")
		o bytes.Buffer
	)

	w, err := crypto.EncryptWriter(x, v, &o)
	if err != nil {
		panic(err)
	}

	w.WriteString("this is a string!")
	w.Flush()
	fmt.Printf("output [%s]\n", o.String())

	r, err := crypto.DecryptReader(x, v, &o)
	if err != nil {
		panic(err)
	}

	c, err := r.StringVal()
	fmt.Printf("input err: %s\ninput out: [%s]\n", err, c)
}
