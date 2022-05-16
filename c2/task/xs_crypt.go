//go:build scripts && crypt

package task

import (
	"github.com/iDigitalFlame/xmt/device/local"
	"github.com/iDigitalFlame/xmt/util/crypt"
)

func createEnvironment() map[string]interface{} {
	return map[string]interface{}{
		crypt.Get(92):  local.UUID.String(),   // ID
		crypt.Get(120): local.Version,         // OS
		crypt.Get(118): local.Device.PID,      // PID
		crypt.Get(119): local.Device.PPID,     // PPID
		crypt.Get(121): local.Elevated(),      // ADMIN
		crypt.Get(122): local.Device.Hostname, // HOSTNAME
	}
}
