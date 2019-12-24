package com

import "testing"

func TestPacket(t *testing.T) {
	p := new(packet)

	t.Logf("%T %v\n", p, p)
}
