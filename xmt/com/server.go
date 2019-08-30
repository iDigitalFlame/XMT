package com

import "github.com/iDigitalFlame/xmt/xmt/device"

type Listener interface {
	Close() error
	String() string
	Accept() (Connection, error)
}

type Connection interface {
	Close() error
	Host() device.ID
	Write(*Packet) error
	Read() (*Packet, error)
}
