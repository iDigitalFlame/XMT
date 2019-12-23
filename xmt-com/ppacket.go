package com

import (
	"io"

	data "github.com/iDigitalFlame/xmt/xmt-data"
	device "github.com/iDigitalFlame/xmt/xmt-device"
)

// PPacket is a work in progress interface for a method to combine packets and streams
// into single
type PPacket interface {
	Reset()
	Clear()
	Len() int
	Size() int
	ID() uint16
	Job() uint16
	Flags() Flag
	IsEmpty() bool
	String() string
	Payload() []byte
	Device() device.ID
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(b []byte) error
	MarshalStream(data.Writer) error
	UnmarshalStream(data.Reader) error
	data.Writer
	data.Reader
	io.Closer
}

type SPacket struct {
	ID, Job uint16
	Device  device.ID
	Single  bool

	blocks      []*block
	last, wrote uint16
}
type block struct {
	f          Flags
	buf        []byte
	rpos, wpos int
}
