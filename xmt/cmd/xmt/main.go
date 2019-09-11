package main

import (
	"crypto/aes"
	"fmt"
	"net/http"
	"time"

	"github.com/iDigitalFlame/logx/logx"
	"github.com/iDigitalFlame/xmt/xmt/com"
	"github.com/iDigitalFlame/xmt/xmt/com/c2"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/tcp"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/transform"
	"github.com/iDigitalFlame/xmt/xmt/com/c2/wrapper"
	"github.com/iDigitalFlame/xmt/xmt/crypto/cbk"
	"github.com/iDigitalFlame/xmt/xmt/crypto/xor"
	"github.com/iDigitalFlame/xmt/xmt/util"
)

func main() {

	testWebC2()
}

func testWebC2() {
	c2.Controller.Log = logx.NewConsole(logx.LDebug)
	c2.Controller.Mux = c2.MuxFunc(func(s *c2.Session, p *com.Packet) {
		fmt.Printf("[%s] Packet: %s\n", s.ID.String(), p.String())
		if p.ID == 910 {
			n1, _ := p.UTFString()
			n2, _ := p.UTFString()
			fmt.Printf("Payload: %s, %s\n", n1, n2)
		}
	})

	aesBlock, _ := aes.NewCipher([]byte("123456789012345678901234567890AB"))
	aesWrapper, _ := wrapper.NewCryptoBlock(aesBlock, []byte("ABCDEF1234567890"))

	cbkWriter := cbk.NewCipher(100)
	cbkWriter.A = 20
	cbkWriter.B = 30
	cbkWriter.C = 40
	cbkReader := cbk.NewCipher(100)
	cbkReader.A = 20
	cbkReader.B = 30
	cbkReader.C = 40

	cbkWrapper, _ := wrapper.NewCrypto(cbkReader, cbkWriter)

	xorCipher := xor.Cipher([]byte("this is my xor key"))
	xorWrapper, _ := wrapper.NewCrypto(xorCipher, xorCipher)

	profile := &c2.Profile{
		Sleep:     time.Duration(5) * time.Second,
		Wrapper:   wrapper.NewMulti(wrapper.Zlib, aesWrapper, cbkWrapper, xorWrapper),
		Transform: transform.Base64,
	}

	generator := &tcp.WebGenerator{
		URL:  util.Matcher("/post/%31fd/%12fd/%31fd/edit/%4fl"),
		Host: util.Matcher("%6fl.myblogsite.com"),
	}

	c2Web := tcp.NewWeb(profile.Sleep)
	c2Web.Handle("/", http.FileServer(http.Dir("/tmp/")))

	c2Web.Rule(&tcp.WebRule{
		URL:  generator.URL.(util.Matcher).Match(),
		Host: generator.Host.(util.Matcher).Match(),
	})

	_, err := c2.Controller.Listen("http", "127.0.0.1:8080", c2Web, profile)
	if err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	c2Client, err := c2.Controller.Connect("http://127.0.0.1:8080", &tcp.WebClient{Generator: generator}, profile)
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	firstPacket := &com.Packet{ID: 910}
	firstPacket.Flags.SetFragTotal(2)
	firstPacket.Flags.SetFragGroup(999)
	firstPacket.Flags.SetFragPosition(0)
	firstPacket.WriteString("testing string 1!")
	fmt.Printf("Sent packet 1: %s\n", firstPacket)
	c2Client.WritePacket(firstPacket)

	time.Sleep(5 * time.Second)

	secondPacket := &com.Packet{ID: 910}
	secondPacket.Flags.SetFragTotal(2)
	secondPacket.Flags.SetFragGroup(999)
	secondPacket.Flags.SetFragPosition(1)
	secondPacket.WriteString("testing string 2!")
	fmt.Printf("Sent packet 2: %s\n", secondPacket)
	c2Client.WritePacket(secondPacket)

	time.Sleep(10 * time.Second)

	c2Client.Shutdown()

	time.Sleep(15 * time.Second)
}
