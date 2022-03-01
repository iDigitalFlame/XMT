//go:build windows
// +build windows

package registry

import (
	"syscall"
	"time"
)

// KeyInfo describes the statistics of a key.
//
// It is returned by a call to Stat.
type KeyInfo struct {
	SubKeyCount     uint32
	MaxSubKeyLen    uint32
	ValueCount      uint32
	MaxValueNameLen uint32
	MaxValueLen     uint32
	lastWriteTime   syscall.Filetime
}

// ModTime returns the key's last write time.
func (i *KeyInfo) ModTime() time.Time {
	return time.Unix(0, i.lastWriteTime.Nanoseconds())
}
