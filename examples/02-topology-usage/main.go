package main

import (
	"git-02.t1-group.ru/go-modules/robbit"
	"os"
)

/*
	This one shows how to declare bindings, queues, exchanges and channels from a configuration file.
*/
func main() {
	topology, _ := os.Open("topology/test-examples/00-simple.yaml")

	c := robbit.To("amqp://localhost:5672/").
		WithTopologyFrom(topology) // ...and that's it.

	// Upon each reconnection it will redeclare the given topology.

	c.RunForever()
}
