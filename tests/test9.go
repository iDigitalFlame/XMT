package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iDigitalFlame/xmt/c2/task"

	"github.com/iDigitalFlame/xmt/c2"
)

func test9Main() {

	var (
		p, err = c2.Config{
			c2.Sleep(time.Second * 5),
			c2.Jitter(10),
			c2.ConnectTCP,
		}.Profile()
	)
	if err != nil {
		panic(err)
	}

	e := make(chan os.Signal, 1)
	signal.Notify(e, syscall.SIGINT)

	go func() {
		fmt.Printf("GOT SIGNAL %s\n", <-e)
		c2.Default.Close()
	}()

	l, err := c2.Listen("test1", "127.0.0.1:8080", nil, p)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listener [%s]\n", l)

	l.New = func(v *c2.Session) {
		d, _ := c2.Default.MarshalJSON()
		fmt.Printf("json payload:\n%s\n", d)
		r, err := v.Schedule(task.Command("ls -al"))
		if err != nil {
			panic(err)
		}
		r.Update = func(j *c2.Job) {
			fmt.Println(j.Status, j.Result)
			d, _ := c2.Default.MarshalJSON()
			fmt.Printf("json payload:\n%s\n", d)
		}

	}

	c, err := c2.Connect("127.0.0.1:8080", nil, p)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Session [%s]\n", c)

	time.Sleep(5 * time.Second)

	d, _ := c2.Default.MarshalJSON()
	fmt.Printf("json payload:\n%s\n", d)

	c2.Default.Wait()
}
