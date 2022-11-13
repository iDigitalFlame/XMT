// Copyright (C) 2020 - 2022 iDigitalFlame
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

package cfg

import (
	"context"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	c := Pack(
		Host("127.0.0.1:8085"),
		ConnectTCP,
		Sleep(time.Second*5),
		Jitter(0),
		Weight(10),
		KillDate(time.Now().AddDate(1, 0, 0)),
	)
	c.AddGroup(
		Host("127.0.0.1:8086"),
		ConnectTCP,
		Sleep(time.Second*10),
		Jitter(50),
		&WorkHours{Days: DayFriday},
	)
	c.AddGroup(
		Host("127.0.0.1:8087"),
		ConnectTCP,
		Sleep(time.Second*5),
		Jitter(0),
	)
	c.AddGroup(Host("127.0.0.1:8088")) // Invalid
	c.Add(SelectorLastValid)
	p, err := Raw(c.Bytes())
	if err != nil {
		t.Fatalf("TestConfig(): Raw failed with error: %s!", err.Error())
	}
	v, ok := p.(*Group)
	if !ok {
		t.Fail()
	}
	if v.Jitter() != 0 {
		t.Fatalf(`TestConfig(): First Jitter should be "0", but is "%d"!`, v.Jitter())
	}
	if h, _, _ := v.Next(); h != "127.0.0.1:8085" {
		t.Fatalf(`TestConfig(): First Host should be "127.0.0.1:8085", but is "%s"!`, h)
	}
	v.Switch(false)
	if v.Jitter() != 0 {
		t.Fatalf(`TestConfig(): First Jitter should be "0", but is "%d"!`, v.Jitter())
	}
	if h, _, _ := v.Next(); h != "127.0.0.1:8085" {
		t.Fatalf(`TestConfig(): First Host should be "127.0.0.1:8085", but is "%s"!`, h)
	}
	v.Switch(true)
	if v.Jitter() != 50 {
		t.Fatalf(`TestConfig(): Second Jitter should be "50", but is "%d"!`, v.Jitter())
	}
	if h, _, _ := v.Next(); h != "127.0.0.1:8086" {
		t.Fatalf(`TestConfig(): Second Host should be "127.0.0.1:8086", but is "%s"!`, h)
	}
	v.Switch(true) // Advance 2 places
	v.Switch(true)
	if _, err := v.Connect(context.Background(), ""); err != ErrNotAConnector {
		t.Fatalf(`TestConfig(): Last Group should raise "ErrNotAConnector" but instead got: %s!`, err)
	}
}
