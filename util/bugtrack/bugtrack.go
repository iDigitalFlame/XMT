//go:build bugs
// +build bugs

package bugtrack

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/PurpleSec/logx"
)

// Enabled is the stats of the bugtrack package.
// This is true if bug tracking is enabled.
const Enabled = true

var log logx.Log

func init() {
	var (
		p   = os.TempDir()
		err = os.MkdirAll(p, 0755)
	)
	if err != nil {
		panic("bugtrack: init failed with error: " + err.Error())
	}
	var (
		f = filepath.Join(p, "bugtrack-"+strconv.Itoa(os.Getpid())+".log")
		l logx.Log
	)
	if l, err = logx.File(f, logx.Append, logx.Trace); err != nil {
		panic("bugtrack: creating file log failed with error: " + err.Error())
	}
	log = logx.Multiple(l, logx.Writer(os.Stderr, logx.Trace))
	log.SetPrefix("BUGTRACK")
	log.Info("Bugtrack log init complete, log file can be found at %q.", f)
}

// Track is a simple logging function that takes the same arguments as a 'fmt.Sprintf'
// function. This can be used to track bugs or output values.
//
// Not recommended to be used in production environments.
// The "-tags bugs" option is required in order for this function to be used.
func Track(s string, m ...interface{}) {
	log.Trace(s, m...)
}
