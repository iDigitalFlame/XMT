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
	"context"
	"fmt"
	"time"
	// You will have to enable the "sotto" package by renaming the "cmd/script/sotto/otto.go_"
	// file to "cmd/script/sotto/otto.go" and run "go mod tidy; go mod verify".
	//
	// The un-comment the line below
	// "github.com/iDigitalFlame/xmt/cmd/script/sotto"
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

func exampleOtto() {
	var (
		l, c   = context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
		r, err = script1, l //  sotto.InvokeContext(l, script1)
	)
	fmt.Printf("output [%s], error[%s]\n", r, err)
	c()
}
