//go:build js || plan9 || aix || illumos || solaris
// +build js plan9 aix illumos solaris

// Copyright (C) 2020 - 2023 iDigitalFlame
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
	"context"
	"syscall"

	"github.com/iDigitalFlame/xmt/data"
)

func taskShutdown(_ context.Context, _ data.Reader, _ data.Writer) error {
	return syscall.EINVAL
}
