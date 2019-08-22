package main

import (
	"bytes"
	"fmt"

	"githuc.com/iDigitalFlame/xmt/xmt/crypto/cbk"
)

func main() {
	b := &bytes.Buffer{}
	w := cbk.NewWriter(cbk.NewCipher(25), b)

	fmt.Printf("%T\n", w)

	n, err := w.Write([]byte("This is a loooooooooooooong string!"))
	fmt.Printf("%d, %s\n", n, err)

	w.Close()

	fmt.Printf("[%s]\n", b.Bytes())

}
