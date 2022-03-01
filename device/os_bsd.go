//go:build aix || dragonfly || freebsd || illumos || netbsd || openbsd || plan9 || solaris || hurd || zos
// +build aix dragonfly freebsd illumos netbsd openbsd plan9 solaris hurd zos

package device

// OS is the local machine's Operating System type.
const OS = Unix
