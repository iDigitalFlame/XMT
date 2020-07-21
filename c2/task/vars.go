package task

import (
	"context"

	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/device"
)

// Custom Task Message ID Values
//
// TvRefresh      - 192:
// TvUpload       - 193:
// TvDownload     - 194:
// TvExecute      - 195:
// TvCode         - 196:
const (
	TvRefresh  uint8 = 0xC0
	TvUpload   uint8 = 0xC1
	TvDownload uint8 = 0xC2
	TvExecute  uint8 = 0xC3
	TvCode     uint8 = 0xC4
)

// Mappings is an fixed size array that contains the Tasker mappings for each ID value. Values that are less than 22
// are ignored. Adding a mapping to here will allow it to be executed via the client Scheduler.
var Mappings = [256]Tasker{
	TvRefresh:  simpleTask(TvRefresh),
	TvUpload:   simpleTask(TvUpload),
	TvDownload: simpleTask(TvDownload),
	TvExecute:  simpleTask(TvExecute),
	TvCode:     simpleTask(TvCode),
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
	case TvCode:
		return code(x, p)
	case TvUpload:
		return upload(x, p)
	case TvRefresh:
		if err := device.Local.Refresh(); err != nil {
			return nil, err
		}
		n := new(com.Packet)
		device.Local.MarshalStream(n)
		return n, nil
	case TvExecute:
		return process(x, p)
	case TvDownload:
		return download(x, p)
	}
	return nil, nil
}
