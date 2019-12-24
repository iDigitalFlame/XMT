package device

import (
	"encoding/json"
	"testing"

	"github.com/iDigitalFlame/xmt/data"
)

func TestDevice(t *testing.T) {
	t.Logf("test")
}

func TestNetwork(t *testing.T) {
	t.Logf("%d Network Interfaces found! Printing details...\n", Local.Network.Len())
	for i := range Local.Network {
		t.Logf("%d: Device: %s\n", i, Local.Network[i].String())
	}

	t.Logf("Attempting to refresh network...\n")
	err := Local.Network.Refresh()
	if err != nil {
		t.Fatalf("Error refreshing network: %s\n", err.Error())
	}

	t.Logf("%d Network Interface found! Printing details...\n", Local.Network.Len())
	for i := range Local.Network {
		t.Logf("%d: Device: %s\n", i, Local.Network[i].String())
	}

	t.Logf("Testing write to binary...\n")
	b := &data.Chunk{}
	err = Local.Network.MarshalStream(b)
	if err != nil {
		t.Fatalf("Error writing to binary: %s\n", err.Error())
	}
	t.Logf("%d bytes written...\n", b.Len())

	t.Logf("Testing read from binary...\n")
	var n Network
	err = n.UnmarshalStream(b)
	if err != nil {
		t.Fatalf("Error reading from binary: %s\n", err.Error())
	}

	t.Logf("%d Devices found in loaded network, printing details...\n", n.Len())
	for i := range n {
		t.Logf("%d: Device: %s\n", i, n[i].String())
	}
	b.Reset()

	if n.Len() != Local.Network.Len() {
		t.Fatalf("Read struct is not equal to written data!")
	}

	t.Logf("Testing write to JSON...\n")
	j, err := json.Marshal(Local.Network)
	if err != nil {
		t.Fatalf("Error writing to JSON: %s\n", err.Error())
	}

	t.Logf("JSON [%s]\n", j)

	t.Logf("Testing read from JSON...\n")
	var m Network
	if err := json.Unmarshal(j, &m); err != nil {
		t.Fatalf("Error writing to JSON: %s\n", err.Error())
	}

	t.Logf("%d Devices found in loaded network, printing details...\n", m.Len())
	for i := range m {
		t.Logf("%d: Device: %s\n", i, m[i].String())
	}

	if m.Len() != Local.Network.Len() {
		t.Fatalf("Read struct is not equal to written data!")
	}
}
