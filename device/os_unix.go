//go:build !windows && !plan9 && !js && !darwin && !linux && !android

package device

// OS is the local machine's Operating System type.
const OS = Unix
