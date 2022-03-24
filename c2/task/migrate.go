package task

import (
	"github.com/iDigitalFlame/xmt/cmd/filter"
	"github.com/iDigitalFlame/xmt/com"
	"github.com/iDigitalFlame/xmt/data"
)

// Callable is an internal interface used to specify a wide range of Runnabale
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

// SpawnPull -
func SpawnPull(f *filter.Filter, s, url string) *com.Packet {
	return SpawnPullProfile(f, s, nil, url)
}

// MigratePull -
func MigratePull(f *filter.Filter, s, url string) *com.Packet {
	return MigratePullProfileEx(f, true, s, nil, url)
}

// Spawn -
func Spawn(f *filter.Filter, s string, c Callable) *com.Packet {
	return SpawnProfile(f, s, nil, c)
}

// Migrate -
func Migrate(f *filter.Filter, s string, c Callable) *com.Packet {
	return MigrateProfileEx(f, true, s, nil, c)
}

// SpawnProfile -
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

// MigrateProfile -
func MigrateProfile(f *filter.Filter, s string, b []byte, c Callable) *com.Packet {
	return MigrateProfileEx(f, true, s, b, c)
}

// SpawnPullProfile -
func SpawnPullProfile(f *filter.Filter, s string, b []byte, url string) *com.Packet {
	n := &com.Packet{ID: MvSpawn}
	n.WriteString(s)
	n.WriteBytes(b)
	f.MarshalStream(n)
	n.WriteUint8(TvPullExecute)
	n.WriteString(url)
	return n
}

// MigratePullProfile -
func MigratePullProfile(f *filter.Filter, s string, b []byte, url string) *com.Packet {
	return MigratePullProfileEx(f, true, s, b, url)
}

// MigrateProfileEx -
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

// MigratePullProfileEx -
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
