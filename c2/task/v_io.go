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
	"io"
	"os"
	"time"

	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

// ProcessList returns a list processes Packet. This can be used to instruct
// the client to return a list of the current running host's processes.
//
// C2 Details:
//
//	ID: MvProcList
//
//	Input:
//	    <none>
//	Output:
//	    uint32          // Count
//	    []ProcessInfo { // List of Running Processes
//	        uint32      // Process ID
//	        uint32      // Parent Process ID
//	        string      // Process Image Name
//	    }
func ProcessList() *com.Packet {
	return &com.Packet{ID: MvProcList}
}

// Kill returns a process kill Packet. This can be used to instruct to send a
// SIGKILL signal to the specified process by the specified Process ID.
//
// C2 Details:
//
//	ID: TvSystemIO
//
//	Input:
//	    uint8  // IO Type
//	    uint32 // PID
//	Output:
//	    uint8  // IO Type
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
//
//	ID: TvSystemIO
//
//	Input:
//	    uint8  // IO Type
//	    string // Path
//	Output:
//	    uint8  // IO Type
func Touch(s string) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoTouch)
	n.WriteString(s)
	return n
}

// KillName returns a process kill Packet. This can be used to instruct to send
// a SIGKILL signal all to the specified processes that have the specified name.
//
// NOTE: This kills all processes that share this name.
//
// C2 Details:
//
//	ID: TvSystemIO
//
//	Input:
//	    uint8  // IO Type
//	    string // Process Name
//	Output:
//	    uint8  // IO Type
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
//
//	ID: TvDownload
//
//	Input:
//	    string // Target
//	Output:
//	    string // Expanded Target Path
//	    bool   // Target is Directory
//	    int64  // Size
//	    []byte // Data
func Download(src string) *com.Packet {
	n := &com.Packet{ID: TvDownload}
	n.WriteString(src)
	return n
}

// Move returns a file move Packet. This can be used to instruct to move the
// specified source file to the specified destination path.
//
// The source and destination paths may contain environment variables that will
// be resolved during runtime.
//
// C2 Details:
//
//	ID: TvSystemIO
//
//	Input:
//	    uint8  // IO Type
//	    string // Source
//	    string // Destination
//	Output:
//	    uint8  // IO Type
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
func Move(src, dst string) *com.Packet {
	n := &com.Packet{ID: TvSystemIO}
	n.WriteUint8(taskIoMove)
	n.WriteString(src)
	n.WriteString(dst)
	return n
}

// Copy returns a file copy Packet. This can be used to instruct to copy the
// specified source file to the specified destination path.
//
// The source and destination paths may contain environment variables that will
// be resolved during runtime.
//
// C2 Details:
//
//	ID: TvSystemIO
//
//	Input:
//	    uint8  // IO Type
//	    string // Source
//	    string // Destination
//	Output:
//	    uint8  // IO Type
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
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
//
//	ID: TvPull
//
//	Input:
//	    string // URL
//	    string // Target Path
//	Output:
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
func Pull(url, path string) *com.Packet {
	return PullAgent(url, "", path)
}

// Upload returns an upload Packet. This will instruct the client to write the
// provided byte array to the filepath provided. The client will return the
// number of bytes written and the resulting expanded file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//
//	ID: TvUpload
//
//	Input:
//	    string // Destination
//	    []byte // File Data
//	Output:
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
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
//
//	ID: TvProcDump
//
//	Input:
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	Output:
//	    []byte // Data
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
//
//	ID: TvSystemIO
//
//	Input:
//	    uint8  // IO Type
//	    string // Path
//	Output:
//	    uint8  // IO Type
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
//
//	ID: TvPull
//
//	Input:
//	    string // URL
//	    string // User-Agent
//	    string // Target Path
//	Output:
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
func PullAgent(url, agent, path string) *com.Packet {
	n := &com.Packet{ID: TvPull}
	n.WriteString(url)
	n.WriteString(agent)
	n.WriteString(path)
	return n
}

// UploadFile returns an upload Packet. This will instruct the client to write
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
//
//	ID: TvUpload
//
//	Input:
//	    string // Destination
//	    []byte // File Data
//	Output:
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
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

// UploadReader returns an upload Packet. This will instruct the client to write
// the provided reader content to the filepath provided. The client will return
// the number of bytes written and the resulting file path.
//
// The destination path may contain environment variables that will be resolved
// during runtime.
//
// C2 Details:
//
//	ID: TvUpload
//
//	Input:
//	    string // Destination
//	    []byte // File Data
//	Output:
//	    string // Expanded Destination Path
//	    uint64 // Byte Count Written
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
//
//	ID: TvPullExecute
//
//	Input:
//	    string          // URL
//	    bool            // Wait
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	Output:
//	    uint32          // PID
//	    int32           // Exit Code
func PullExecute(url string, w bool, f *filter.Filter) *com.Packet {
	return PullExecuteAgent(url, "", w, f)
}

// Restart returns a shutdown Packet. This will instruct the client to initiate
// a restart/reboot operation. A reboot message, reason, force and timeout can
// be specified. Timeouts are specified in seconds.
//
// Message and Reason codes are only applicable to Windows devices and are ignored
// on non-Windows devices.
//
// C2 Details:
//
//	ID: TvPower
//
//	Input:
//	    string // Restart message (Windows only)
//	    uint32 // Timeout (seconds)
//	    uint32 // Reason code (Windows only)
//	    uint8  // Flags
//	Output:
//	    <none>
func Restart(msg string, sec uint32, force bool, reason uint32) *com.Packet {
	n := &com.Packet{ID: TvPower}
	n.WriteString(msg)
	n.WriteUint32(sec)
	n.WriteUint32(reason)
	var v uint8 = 1
	if force {
		v |= 2
	}
	n.WriteUint8(v)
	return n
}

// Shutdown returns a shutdown Packet. This will instruct the client to initiate
// a shutdown/poweroff operation. A shutdown message, reason, force and timeout
// can be specified. Timeouts are specified in seconds.
//
// Message and Reason codes are only applicable to Windows devices and are ignored
// on non-Windows devices.
//
// C2 Details:
//
//	ID: TvPower
//
//	Input:
//	    string // Shutdown message (Windows only)
//	    uint32 // Timeout (seconds)
//	    uint32 // Reason code (Windows only)
//	    uint8  // Flags
//	Output:
//	    <none>
func Shutdown(msg string, sec uint32, force bool, reason uint32) *com.Packet {
	n := &com.Packet{ID: TvPower}
	n.WriteString(msg)
	n.WriteUint32(sec)
	n.WriteUint32(reason)
	var v uint8 = 0
	if force {
		v |= 2
	}
	n.WriteUint8(v)
	return n
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
//
//	ID: TvPullExecute
//
//	Input:
//	    string          // URL
//	    string          // User-Agent
//	    bool            // Wait
//	    Filter struct { // Filter
//	        bool        // Filter Status
//	        uint32      // PID
//	        bool        // Fallback
//	        uint8       // Session
//	        uint8       // Elevated
//	        []string    // Exclude
//	        []string    // Include
//	    }
//	Output:
//	    uint32          // PID
//	    int32           // Exit Code
func PullExecuteAgent(url, agent string, w bool, f *filter.Filter) *com.Packet {
	n := &com.Packet{ID: TvPullExecute}
	n.WriteString(url)
	n.WriteString(agent)
	n.WriteBool(w)
	f.MarshalStream(n)
	return n
}

// Netcat returns a network connection Packet. This will instruct the client to
// initiate a network call to the specified host:port with the provided protocol.
// Reading the results and timeouts can be specified, along with the payload to
// be sent.
//
// If 'read' is true, the resulting data stream results would be returned.
//
// C2 Details:
//
//	ID: TvNetcat
//
//	Input:
//	    string // Host:Port
//	    uint8  // Read | Protocol
//	    uint64 // Timeout
//	    []byte // Data to send
//	Output:
//	    []byte // Result data (if read is true)
func Netcat(host string, proto uint8, t time.Duration, read bool, b []byte) *com.Packet {
	n := &com.Packet{ID: TvNetcat}
	n.WriteString(host)
	p := proto
	if read {
		p |= 128
	}
	n.WriteUint8(p)
	n.WriteUint64(uint64(t))
	n.WriteBytes(b)
	return n
}
