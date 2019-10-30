package main

import (
	"github.com/streadway/amqp"
)
import "git-02.t1-group.ru/go-modules/robbit"

/*
	If initialization of at least one channel declared with MaintainChannel panics, the whole connection goes reconnect.
	This example reconnects infinitely.
*/
func main() {
	c := robbit.ConnectTo("amqp://localhost:5672/")

	c.MaintainChannel("source", func(channel *amqp.Channel, connection *amqp.Connection) {
		fmt.Println("Channel", channel, "is given")
		panic("me ded lol")
	})

	c.RunForever()
}
