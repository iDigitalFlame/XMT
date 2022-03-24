package main

import (
	"context"
	"fmt"
	"time"
	//"github.com/iDigitalFlame/xmt/cmd/script/sotto"
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

func testOtto() {
	var (
		l, c   = context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
		r, err = "", l //  sotto.InvokeContext(l, script1)
	)

	fmt.Printf("output [%s], error[%s]\n", r, err)
	c()
}
