//go:build scripts && crypt

package task

import (
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func createEnvironment() map[string]any {
	return map[string]any{
		crypt.Get(8):  local.UUID.String(),   // ID
		crypt.Get(9):  local.Version,         // OS
		crypt.Get(10): local.Device.PID,      // PID
		crypt.Get(11): local.Device.PPID,     // PPID
		crypt.Get(12): local.Elevated(),      // ADMIN
		crypt.Get(13): local.Device.Hostname, // HOSTNAME
	}
}
