package main

import (
	"fmt"

	"github.com/iDigitalFlame/xmt/util"
)

func main() {
	var (
		m = util.Matcher(
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
