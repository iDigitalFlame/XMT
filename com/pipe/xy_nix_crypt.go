//go:build !windows && crypt

package pipe

import (
	"os"
	"path/filepath"

	"github.com/iDigitalFlame/xmt/util/crypt"
)

// PermEveryone is the Linux permission string used in sockets to allow anyone to write and read
// to the listening socket. This can be used for socket communcation between privilege boundaries.
// This can be applied to the ListenPerm function.
var PermEveryone = crypt.Get(36) // 0766

// Format will ensure the path for this Pipe socket fits the proper OS based pathname. Valid pathnames will be
// returned without any changes.
func Format(s string) string {
	if !filepath.IsAbs(s) {
		var (
			p      = crypt.Get(37) + s                          // /run/
			f, err = os.OpenFile(crypt.Get(37)+s, 0x242, 0o400) // /run/
			// 0x242 - CREATE | TRUNCATE | RDWR
		)
		if err != nil {
			return crypt.Get(38) + s // /tmp/
		}
		f.Close()
		os.Remove(p)
		return p
	}
	return s
}
