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

package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/iDigitalFlame/xmt/com/pipe"
)

func examplePipes() {
	l, err := pipe.ListenPerms(pipe.Format("testing1"), pipe.PermEveryone)
	if err != nil {
		panic(err)
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-s
		l.Close()
	}()

	for {
		c, err := l.Accept()
		if err != nil {
			break
		}
		go io.Copy(os.Stdout, c)
	}

	close(s)
}
