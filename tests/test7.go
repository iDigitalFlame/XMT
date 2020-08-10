package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iDigitalFlame/xmt/cmd/script"
	"github.com/iDigitalFlame/xmt/man"
	"github.com/iDigitalFlame/xmt/util/text"
)

const script1 = `
// JavaScript test via Otto.
//
function derp() {
	console.log("hello!");
}

derp();
console.log(exec('pwd'));

var a = 2;
var v = exec("bash -c 'echo \"derp " + a + " value\" > /tmp/derp.txt'");
console.log("got " + v);

print("Sleeping 5 seconds..");
sleep(0.5);
print("Done!");

var f = exec("cat /tmp/derp.txt");
print("read " + f, f);
exec("rm /tmp/derp.txt");
`

func test7Main() {
	var (
		l, c   = context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
		r, err = script.InvokeOttoContext(l, script1)
	)

	fmt.Printf("output [%s], error[%s]\n", r, err)
	c()
}

func test7Main2() {
	k := man.MustGuard("testing12")
	defer k.Close()

	var (
		m = text.Matcher(
			"test1: %d %5d %5fd %h %h %5h %5fh %n %5n %5fn %c %5c %5fc %u %5u %5fu %l %5l %5fl %s %5s %5fs",
		)
		x = m.Match()
	)
	for i := 0; i < 100; i++ {
		v := m.String()
		if !x.MatchString(v) {
			panic(fmt.Sprintf("string [%s] did not match [%s]!", v, x))
		}
		fmt.Printf("%s - Matched!\n", v)
	}
}
