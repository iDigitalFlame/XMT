package main

import (
	"bytes"
	"fmt"

	"./xmt/device"
	"./xmt/dio"
)

func main() {

	fmt.Printf("%v\n", device.Local)

	b := &bytes.Buffer{}
	w := dio.NewWriter(b)

	w.WriteString("derp")
	w.WriteString("lolcopter")
	w.WriteInt32(int32(978654321))
	w.WriteInt16(int16(32101))
	w.WriteInt64(int64(9012345678901))
	w.WriteUint16(uint16(65432))
	w.WriteString("This is a weird string, hello there!!")

	device.Local.MarshalStream(w)

	if err := w.Close(); err != nil {
		panic(err)
	}

	o := b.Bytes()
	fmt.Printf("Results\nRaw: [%v]\nString: [%s]\n", o, o)

	r := dio.NewByteReader(o)

	var r1, r2, r3 string
	var r4 int16
	var r5 int32
	var r6 int64
	var r7 uint16

	if err := r.ReadString(&r1); err != nil {
		panic(err)
	}
	if err := r.ReadString(&r2); err != nil {
		panic(err)
	}
	if err := r.ReadInt32(&r5); err != nil {
		panic(err)
	}
	if err := r.ReadInt16(&r4); err != nil {
		panic(err)
	}
	if err := r.ReadInt64(&r6); err != nil {
		panic(err)
	}
	if err := r.ReadUint16(&r7); err != nil {
		panic(err)
	}
	if err := r.ReadString(&r3); err != nil {
		panic(err)
	}

	fmt.Printf("After: r1: (%s) r2: (%s), r3: (%s), r4: (%d), r5: (%d), r6: (%d), r7: (%d)\n", r1, r2, r3, r4, r5, r6, r7)
}
