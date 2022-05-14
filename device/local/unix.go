//go:build !windows && !js

package local

import (
	"os"
	"os/user"
)

func isElevated() uint8 {
	if os.Geteuid() == 0 || os.Getuid() == 0 {
		return 1
	}
	if a, err := user.Current(); err == nil && a.Uid == "0" {
		return 1
	}
	return 0
}
