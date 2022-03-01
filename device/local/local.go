package local

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"os"
	"os/user"

	"github.com/iDigitalFlame/xmt/device"
)

var (
	// UUID is the device specific and session specific identifier.
	UUID = getID()

	// Version is the local machine's Operating System version information.
	Version = version()
)

// Device is the pointer to the local machine instance. This instance is loaded
// at runtime and is used for local data gathering and identification.
var Device = (&local{&device.Machine{
	ID:       UUID,
	OS:       device.OS,
	PID:      uint32(os.Getpid()),
	Arch:     device.Arch,
	PPID:     uint32(os.Getppid()),
	Version:  Version,
	Network:  make(device.Network, 0),
	Elevated: isElevated(),
}}).init()

type local struct {
	*device.Machine
}

// Elevated will return true if the current process has elevated privileges,
// false otherwise.
//
// This function is evaluated at runtime.
func Elevated() bool {
	return Device.Elevated
}
func getID() device.ID {
	var i device.ID
	if s := sysID(); len(s) > 0 {
		h := hmac.New(sha256.New, s)
		h.Write([]byte(vers))
		copy(i[:], h.Sum(nil))
		h.Reset()
	} else {
		rand.Read(i[:])
	}
	// NOTE(dij): ID changes from v3 => v4
	//             - Dropped the "machineid" library
	//             - Windows now uses the system SID instead
	//               - Falls back to machine GUID if that fails
	//             - Fixes for AIX and BSD UUID grab
	//             - Fixed an iOS/MacOS ID pickup bug
	//
	//            This code below changes how ID's are generated
	//            An extra bit is taken away from the random address space
	//            (8 => 7), thus short IDs from the same machine will ALWAYS
	//            have the same two first bits for easy identification.
	if rand.Read(i[device.MachineIDSize+1:]); i[0] == 0 {
		i[0] = 1
	}
	if i[device.MachineIDSize] == 0 {
		i[device.MachineIDSize] = 1
	}
	return i
}
func (l *local) init() *local {
	if u, err := user.Current(); err == nil {
		l.User = u.Username
	}
	l.Hostname, _ = os.Hostname()
	l.Network.Refresh()
	l.fixHostname()
	return l
}
func (l *local) fixHostname() {
	for c, i := 0, 0; i < len(l.Hostname); i++ {
		if l.Hostname[i] != '.' {
			continue
		}
		if c++; c > 1 {
			l.Hostname = l.Hostname[:i]
			return
		}
	}
}
func (l *local) Refresh() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	l.User = u.Username
	if l.Hostname, err = os.Hostname(); err != nil {
		return err
	}
	if err := l.Network.Refresh(); err != nil {
		return err
	}
	l.PID = uint32(os.Getpid())
	l.PPID = uint32(os.Getppid())
	l.Elevated = isElevated()
	l.fixHostname()
	return nil
}
