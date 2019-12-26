package data

import (
	"bytes"
	"testing"
)

func TestChunk(t *testing.T) {
	c := new(Chunk)

	if err := c.WriteString("testing string"); err != nil {
		t.Fatalf("failed writing test string: %s\n", err.Error())
	}

	var (
		resultStr = "testing string"
		resultBin = []byte{1, 16, 1, 14, 116, 101, 115, 116, 105, 110, 103, 32, 115, 116, 114, 105, 110, 103}
	)

	var (
		b = new(bytes.Buffer)
		w = NewWriter(b)
	)

	if err := c.MarshalStream(w); err != nil {
		t.Fatalf("failed marshaling to binary: %s\n", err.Error())
	}

	var (
		n = new(Chunk)
		d = b.Bytes()
		r = NewReader(bytes.NewReader(d))
	)

	if !bytes.Equal(d, resultBin) {
		t.Fatalf("marsheled bytes do not match expected: got %v, expected %v\n", d, resultBin)
	}

	if err := n.UnmarshalStream(r); err != nil {
		t.Fatalf("failed unmarshaling from binary: %s\n", err.Error())
	}

	s, err := n.StringVal()
	if err != nil {
		t.Fatalf("failed reading from new Chunk: %s\n", err.Error())
	}

	if s != resultStr {
		t.Fatalf("read string does not match expected: got %s, expected %s\n", s, resultStr)
	}

}
