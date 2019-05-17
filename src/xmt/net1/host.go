package net

import "xmt/device"

type Host interface {
	Hostname() string
	Addresss() string
	Os() *device.OsDetails
	UnmarshalBinary([]byte) error
	MarshalBinary() ([]byte, error)
}
