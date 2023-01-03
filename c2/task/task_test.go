//go:build !implant

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
	"testing"
	"time"

	"github.com/iDigitalFlame/xmt/com"
)

func TestDLL(t *testing.T) {
	p, err := DLL{Data: []byte("hello"), Timeout: time.Hour}.Packet()
	if err != nil {
		t.Fatalf("TestDLL(): Packet returned error: %s!", err.Error())
	}
	if n := p.Size(); n <= com.PacketHeaderSize {
		t.Fatalf("TestDLL(): Packet result was empty!")
	}
}
func TestTasks(t *testing.T) {
	v := [...]struct {
		Packet *com.Packet
		Size   int
		ID     uint8
	}{
		{FuncUnmapAll(), 1, TvFuncMap},
		{Evade(0), 1, TvEvade},
		{FuncRemapList(), 0, TvFuncMapList},
		{FuncUnmap("test"), 5, TvFuncMap},
		{CheckDLLFile("file"), 12, TvCheck},
		{PatchDLLFile("file"), 12, TvPatch},
		{FuncRemap("test", []byte("test")), 11, TvFuncMap},
		{CheckFunction("test", "test", []byte("test")), 22, TvCheck},
		{PatchFunction("test", "test", []byte("test")), 22, TvPatch},
		{CheckDLL("test", 0x1234, []byte("test")), 17, TvCheck},
		{PatchDLL("test", 0x1234, []byte("test")), 17, TvPatch},
		{CheckFunctionFile("test", "test", []byte("test")), 22, TvCheck},
		{ProcessList(), 0, MvProcList},
		{Kill(0x2345), 5, TvSystemIO},
		{Touch("test"), 7, TvSystemIO},
		{KillName("test"), 7, TvSystemIO},
		{Download("test"), 6, TvDownload},
		{Move("test", "test"), 13, TvSystemIO},
		{Copy("test", "test"), 13, TvSystemIO},
		{Pull("test", "test"), 13, TvPull},
		{Upload("test", []byte("test")), 10, TvUpload},
		{Delete("test", true), 7, TvSystemIO},
		{PullAgent("test", "test", "test"), 18, TvPull},
		{Restart("test", 0, true, 0x5), 15, TvPower},
		{Shutdown("test", 0, true, 0x5), 15, TvPower},
		{Netcat("test", 0, 0, false, []byte("test")), 15, TvNetcat},
		{RegLs("test"), 7, TvRegistry},
		{RegMakeKey("test"), 7, TvRegistry},
		{RegGet("test", "test"), 13, TvRegistry},
		{RegSetString("test", "test", "test"), 19, TvRegistry},
		{RegDeleteKey("test", true), 8, TvRegistry},
		{RegDelete("test", "test", false), 13, TvRegistry},
		{RegSetDword("test", "test", 0xBEEF), 17, TvRegistry},
		{RegSetQword("test", "test", 0xDEADBEEF), 21, TvRegistry},
		{RegSetBytes("test", "test", []byte("test")), 19, TvRegistry},
		{RegSetExpandString("test", "test", "test"), 19, TvRegistry},
		{RegSet("test", "test", 0, []byte("test")), 23, TvRegistry},
		{RegSetStringList("test", "test", []string{"test"}), 21, TvRegistry},
		{Pwd(), 0, MvPwd},
		{Mounts(), 0, MvMounts},
		{Refresh(), 0, MvRefresh},
		{RevToSelf(), 0, TvRevSelf},
		{UserLogins(), 0, TvLogins},
		{ScreenShot(), 0, TvScreenShot},
		{Ls("test"), 6, MvList},
		{IsDebugged(), 0, MvCheckDebug},
		{Jitter(10), 10, MvTime},
		{Cwd("test"), 6, MvCwd},
		{Profile([]byte("test")), 6, MvProfile},
		{KillDate(time.Time{}), 9, MvTime},
		{ProcessName("test"), 6, TvRename},
		{Wait(time.Hour), 8, TvWait},
		{UserLogoff(0), 5, TvLoginsAct},
		{Sleep(time.Minute), 10, MvTime},
		{ProxyRemove("test"), 7, MvProxy},
		{UserProcesses(1), 4, TvLoginsProc},
		{UserDisconnect(2), 5, TvLoginsAct},
		{Duration(time.Hour, 10), 10, MvTime},
		{Proxy("test", "test", []byte("test")), 19, MvProxy},
		{ProxyReplace("test", "test", []byte("test")), 19, MvProxy},
		{LoginUser(false, "test", "test", "test"), 19, TvLoginUser},
		{WorkHours(1, 2, 3, 4, 5), 6, MvTime},
		{WindowList(), 0, TvWindowList},
		{SwapMouse(false), 1, TvTroll},
		{BlockInput(true), 1, TvTroll},
		{Wallpaper("test"), 7, TvTroll},
		{HighContrast(true), 1, TvTroll},
		{WindowFocus(0xBEEF), 9, TvUI},
		{WindowClose(0xBEEF), 9, TvUI},
		{WallpaperBytes([]byte("test")), 5, TvTroll},
		{WindowWTF(time.Hour), 9, TvTroll},
		{WindowShow(0xBEEF, 0), 10, TvUI},
		{WindowEnable(0xBEEF, true), 9, TvUI},
		{WindowSendInput(0xBEEF, "test"), 15, TvUI},
		{WindowTransparency(0xBEEF, 127), 10, TvUI},
		{WindowMove(0xBEEF, 1, 2, 3, 4), 25, TvUI},
		{WindowMessageBox(0xBEEF, "test", "test", 123), 25, TvUI},
		{UserMessageBox(0, "test", "test", 1, 2, true), 26, TvLoginsAct},
	}
	for i := range v {
		if v[i].Packet == nil {
			t.Fatalf(`TestPacket(): Packet index "%d" was nil!`, i)
		}
		if v[i].ID != v[i].Packet.ID {
			t.Fatalf(`TestPacket(): Packet ID "%d" does not match expected ID "%d"!`, v[i].Packet.ID, v[i].ID)
		}
		if n := v[i].Packet.Size(); n < v[i].Size+com.PacketHeaderSize {
			t.Fatalf(`TestPacket(): Packet size "%d" does not match the expected size "%d"!`, n, v[i].Size+com.PacketHeaderSize)
		}
	}
}
func TestZombie(t *testing.T) {
	p, err := Zombie{Data: []byte("hello"), Timeout: time.Hour}.Packet()
	if err != nil {
		t.Fatalf("TestZombie(): Packet returned error: %s!", err.Error())
	}
	if n := p.Size(); n <= com.PacketHeaderSize {
		t.Fatalf("TestZombie(): Packet result was empty!")
	}
}
func TestProcess(t *testing.T) {
	p, err := Process{Args: []string{"test1", "test2"}, User: "bob", Pass: "password", Timeout: time.Hour}.Packet()
	if err != nil {
		t.Fatalf("TestProcess(): Packet returned error: %s!", err.Error())
	}
	if n := p.Size(); n <= com.PacketHeaderSize {
		t.Fatalf("TestProcess(): Packet result was empty!")
	}
	p.Seek(0, 0)
	a, _, err := ProcessUnmarshal(context.Background(), p)
	if err != nil {
		t.Fatalf("TestProcess(): AssemblyUnmarshal() returned error: %s!", err.Error())
	}
	if a.Args[0] != "test1" || a.Args[1] != "test2" {
		t.Fatalf(`TestProcess(): Args "%s" did not equal "[test1, test2]"!`, a.Args)
	}
	if a.Timeout != time.Hour {
		t.Fatalf(`TestProcess(): Timeout "%s" did not equal "1h"!`, a.Timeout.String())
	}
}
func TestAssembly(t *testing.T) {
	p, err := Assembly{Data: []byte("hello"), Timeout: time.Hour}.Packet()
	if err != nil {
		t.Fatalf("TestAssembly(): Packet returned error: %s!", err.Error())
	}
	if n := p.Size(); n <= com.PacketHeaderSize {
		t.Fatalf("TestAssembly(): Packet result was empty!")
	}
	p.Seek(0, 0)
	a, _, err := AssemblyUnmarshal(context.Background(), p)
	if err != nil {
		t.Fatalf("TestAssembly(): AssemblyUnmarshal() returned error: %s!", err.Error())
	}
	if string(a.Data) != "hello" {
		t.Fatalf(`TestAssembly(): Data "%s" did not equal "hello"!`, a.Data)
	}
	if a.Timeout != time.Hour {
		t.Fatalf(`TestAssembly(): Timeout "%s" did not equal "1h"!`, a.Timeout.String())
	}
}
