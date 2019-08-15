package robbit

import (
	"io/ioutil"
	"testing"
)

func TestDecompose(t *testing.T) {
	data, _ := ioutil.ReadFile("./topology/test-examples/00-simple.yaml")

	res := DecomposeTopology(string(data))

	if len(res.Channels) != 3 {
		t.Error("Strange number of channels")
	}

	if len(res.Queues) != 1 {
		t.Error("Strange number of queues")
	}

	if len(res.Exchanges) != 1 {
		t.Error("Strange number of exchanges")
	}

	if len(res.Bindings) != 1 {
		t.Error("Strange number of bindings")
	}

	if res.ChannelForDeclarations != "lol" {
		t.Error("Strange channel for declarations")
	}
}
