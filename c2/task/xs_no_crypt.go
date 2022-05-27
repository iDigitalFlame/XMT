//go:build scripts && !crypt

package task

import "github.com/iDigitalFlame/xmt/device/local"

func createEnvironment() map[string]any {
	return map[string]any{
		"ID":       local.UUID.String(),
		"OS":       local.Version,
		"PID":      local.Device.PID,
		"PPID":     local.Device.PPID,
		"ADMIN":    local.Elevated(),
		"HOSTNAME": local.Device.Hostname,
	}
}
