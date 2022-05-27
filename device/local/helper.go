//go:build !windows && !js && !linux && !android && crypt

package local

import (
	"os/exec"
	"strings"
)

func output(s string) ([]byte, error) {
	return (&exec.Cmd{Args: strings.Split(s, " ")}).CombinedOutput()
}
