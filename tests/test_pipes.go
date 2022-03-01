package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/iDigitalFlame/xmt/com/pipe"
)

func testPipes() {
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
