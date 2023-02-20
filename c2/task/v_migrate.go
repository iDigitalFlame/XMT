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

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

var (
	_ Callable = (*DLL)(nil)
	_ Callable = (*Zombie)(nil)
	_ Callable = (*Process)(nil)
	_ Callable = (*Assembly)(nil)
)

// Callable is an internal interface used to specify a wide range of Runnable
// types that can be Marshaled into a Packet.
//
// Currently the DLL, Zombie, Assembly and Process instances are supported.
type Callable interface {
	task() uint8
	MarshalStream(data.Writer) error
}

func (DLL) task() uint8 {
	return TvDLL
}
func (Zombie) task() uint8 {
	return TvZombie
}
func (Process) task() uint8 {
	return TvExecute
}
func (Assembly) task() uint8 {
	return TvAssembly
}

// Spawn will attempt to spawn a new instance using the provided Callable type
// as the source.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// The return result is the PID of the new instance.
//
// This function uses the same Profile as the target Session. Use the
// 'SpawnProfile' function to change this behavior.
//
// C2 Details:
//
//	ID: MvSpawn
//
//	Input:
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type
//	    <...>           // Callable Data
//	Output:
//	    uint32          // New PID
func Spawn(f *filter.Filter, s string, c Callable) *com.Packet {
	return SpawnProfile(f, s, nil, c)
}

// Migrate will attempt to migrate to a new instance using the provided Callable
// type as the source.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// This function keeps the same Profile. Use the 'MigrateProfile' or
// 'MigrateProfileEx' function to change this behavior.
//
// This function will automatically wait for all Jobs to complete. Use the
// 'MigrateProfileEx' function to change this behavior.
//
// C2 Details:
//
//	ID: MvMigrate
//
//	Input:
//	    bool            // Wait for Jobs
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type
//	    <...>           // Callable Data
//	Output:
//	    <none>          // RvResult packet sent separately
func Migrate(f *filter.Filter, s string, c Callable) *com.Packet {
	return MigrateProfileEx(f, true, s, nil, c)
}

// SpawnPull will attempt to spawn a new instance using the provided URL as the
// source.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// The return result is the PID of the new instance.
//
// This function uses the same Profile as the target Session. Use the
// 'SpawnPullProfile' function to change this behavior.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension. (see 'man.ParseDownloadHeader')
//
// C2 Details:
//
//	ID: MvSpawn
//
//	Input:
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type (always TvPullExecute)
//	    string          // URL
//	    string          // User-Agent
//	Output:
//	    uint32          // New PID
func SpawnPull(f *filter.Filter, s, url, agent string) *com.Packet {
	return SpawnPullProfile(f, s, nil, url, agent)
}

// MigratePull will attempt to migrate to a new instance using the provided URL
// as the source.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// This function keeps the same Profile. Use the 'MigratePullProfile' or
// 'MigratePullProfileEx' function to change this behavior.
//
// This function will automatically wait for all Jobs to complete. Use the
// 'MigratePullProfileEx' function to change this behavior.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension. (see 'man.ParseDownloadHeader')
//
// C2 Details:
//
//	ID: MvMigrate
//
//	Input:
//	    bool            // Wait for Jobs
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type (always TvPullExecute)
//	    string          // URL
//	    string          // User-Agent
//	Output:
//	    <none>          // RvResult packet sent separately
func MigratePull(f *filter.Filter, s, url, agent string) *com.Packet {
	return MigratePullProfileEx(f, true, s, nil, url, agent)
}

// SpawnProfile will attempt to spawn a new instance using the provided Callable
// type as the source with the supplied Profile bytes.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// The return result is the PID of the new instance.
//
// If the 'b' Profile bytes is nil or empty, the current target Session Profile
// will be used.
//
// C2 Details:
//
//	ID: MvSpawn
//
//	Input:
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type
//	    <...>           // Callable Data
//	Output:
//	    uint32          // New PID
func SpawnProfile(f *filter.Filter, s string, b []byte, c Callable) *com.Packet {
	n := &com.Packet{ID: MvSpawn}
	n.WriteString(s)
	n.WriteBytes(b)
	if f.MarshalStream(n); c == nil {
		n.WriteUint8(0)
		return n
	}
	n.WriteUint8(c.task())
	c.MarshalStream(n)
	return n
}

// MigrateProfile will attempt to migrate to a new instance using the provided
// Callable type as the source with the supplied Profile bytes.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// If the 'b' Profile bytes is nil or empty, the current target Session Profile
// will be used.
//
// This function will automatically wait for all Jobs to complete. Use the
// 'MigrateProfileEx' function to change this behavior.
//
// C2 Details:
//
//	ID: MvMigrate
//
//	Input:
//	    bool            // Wait for Jobs
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type
//	    <...>           // Callable Data
//	Output:
//	    <none>          // RvResult packet sent separately
func MigrateProfile(f *filter.Filter, s string, b []byte, c Callable) *com.Packet {
	return MigrateProfileEx(f, true, s, b, c)
}

// SpawnPullProfile will attempt to spawn a new instance using the provided URL
// as the source with the supplied Profile bytes.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// The return result is the PID of the new instance.
//
// If the 'b' Profile bytes is nil or empty, the current target Session Profile
// will be used.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension. (see 'man.ParseDownloadHeader')
//
// C2 Details:
//
//	ID: MvSpawn
//
//	Input:
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type (always TvPullExecute)
//	    string          // URL
//	    string          // User-Agent
//	Output:
//	    uint32          // New PID
func SpawnPullProfile(f *filter.Filter, s string, b []byte, url, agent string) *com.Packet {
	n := &com.Packet{ID: MvSpawn}
	n.WriteString(s)
	n.WriteBytes(b)
	f.MarshalStream(n)
	n.WriteUint8(TvPullExecute)
	n.WriteString(url)
	n.WriteString(agent)
	return n
}

// MigrateProfileEx will attempt to migrate to a new instance using the provided
// Callable type as the source with the supplied Profile bytes and the 'w' boolean
// to specify waiting for Jobs to complete.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// If the 'b' Profile bytes is nil or empty, the current target Session Profile
// will be used.
//
// C2 Details:
//
//	ID: MvMigrate
//
//	Input:
//	    bool            // Wait for Jobs
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type
//	    <...>           // Callable Data
//	Output:
//	    <none>          // RvResult packet sent separately
func MigrateProfileEx(f *filter.Filter, w bool, s string, b []byte, c Callable) *com.Packet {
	n := &com.Packet{ID: MvMigrate}
	n.WriteBool(w)
	n.WriteString(s)
	n.WriteBytes(b)
	if f.MarshalStream(n); c == nil {
		n.WriteUint8(0)
		return n
	}
	n.WriteUint8(c.task())
	c.MarshalStream(n)
	return n
}

// MigratePullProfile will attempt to migrate to a new instance using the
// provided URL as the source with the supplied Profile bytes.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// If the 'b' Profile bytes is nil or empty, the current target Session Profile
// will be used.
//
// This function will automatically wait for all Jobs to complete. Use the
// 'MigratePullProfileEx' function to change this behavior.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension. (see 'man.ParseDownloadHeader')
//
// C2 Details:
//
//	ID: MvMigrate
//
//	Input:
//	    bool            // Wait for Jobs
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type (always TvPullExecute)
//	    string          // URL
//	    string          // User-Agent
//	Output:
//	    <none>          // RvResult packet sent separately
func MigratePullProfile(f *filter.Filter, s string, b []byte, url, agent string) *com.Packet {
	return MigratePullProfileEx(f, true, s, b, url, agent)
}

// MigratePullProfileEx will attempt to migrate to a new instance using the
// provided URL as the source with the supplied Profile bytes and the 'w' boolean
// to specify waiting for Jobs to complete.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The provided Filter specifies the parent of the new instance and the 's'
// argument string specifies the pipe name to use while connecting.
//
// If the 'b' Profile bytes is nil or empty, the current target Session Profile
// will be used.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension. (see 'man.ParseDownloadHeader')
//
// C2 Details:
//
//	ID: MvMigrate
//
//	Input:
//	    bool            // Wait for Jobs
//	    string          // Pipe Name
//	    []byte          // Profile Bytes
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	    uint8           // Callable Type (always TvPullExecute)
//	    string          // URL
//	    string          // User-Agent
//	Output:
//	    <none>          // RvResult packet sent separately
func MigratePullProfileEx(f *filter.Filter, w bool, s string, b []byte, url, agent string) *com.Packet {
	n := &com.Packet{ID: MvMigrate}
	n.WriteBool(w)
	n.WriteString(s)
	n.WriteBytes(b)
	f.MarshalStream(n)
	n.WriteUint8(TvPullExecute)
	n.WriteString(url)
	n.WriteString(agent)
	return n
}
