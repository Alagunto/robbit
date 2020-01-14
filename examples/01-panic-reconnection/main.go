package main

import (
	"fmt"
	"github.com/alagunto/robbit"
)

/*
	If initialization of at least one channel declared with MaintainChannel panics, the whole connection goes reconnect.
	This example reconnects infinitely.
*/
func main() {
	c := robbit.To("amqp://localhost:5672/")

	c.MaintainChannel("source", func(connection *robbit.Connection, channel *robbit.Channel) {
		fmt.Println("Channel", channel.Key, "is given")
		panic("me ded lol")
	})

	c.RunForever()
}
