package cmd

import "github.com/iDigitalFlame/xmt/data"

// ProcessInfo is a struct that holds simple process related data for extraction
// This struct is returned via a call to 'Processes'.
//
// This struct also supports binary Marshaling/UnMarshaling.
type ProcessInfo struct {
	Name      string
	PID, PPID uint32
}
type processList []ProcessInfo

func (p processList) Len() int {
	return len(p)
}
func (p processList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p processList) Less(i, j int) bool {
	if p[i].PPID == p[j].PPID {
		return p[i].PID < p[j].PID
	}
	return p[i].PPID < p[j].PPID
}

// MarshalStream transforms this struct into a binary format and writes to the
// supplied data.Writer.
func (p ProcessInfo) MarshalStream(w data.Writer) error {
	if err := w.WriteUint32(p.PID); err != nil {
		return err
	}
	if err := w.WriteUint32(p.PPID); err != nil {
		return err
	}
	if err := w.WriteString(p.Name); err != nil {
		return err
	}
	return nil
}

// UnmarshalStream transforms this struct from a binary format that is read from
// the supplied data.Reader.
func (p *ProcessInfo) UnmarshalStream(r data.Reader) error {
	if err := r.ReadUint32(&p.PID); err != nil {
		return err
	}
	if err := r.ReadUint32(&p.PPID); err != nil {
		return err
	}
	if err := r.ReadString(&p.Name); err != nil {
		return err
	}
	return nil
}
