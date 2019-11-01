package main

import (
	"fmt"
)
import "git-02.t1-group.ru/go-modules/robbit"

func main() {
	// It declares a new connection, not actually connecting yet
	c := robbit.To("amqp://localhost:5672/")

	// It tells connection to create a channel each time it's reconnecting.
	// After each channel creation, given callback will be run
	c.MaintainChannel("source", func(connection *robbit.Connection, channel *robbit.Channel) {
		fmt.Println("Channel", channel.Key, "is given")
	})

	c.InitializeWith(func(connection *robbit.Connection) {
		for c, _ := range connection.OpenChannels {
			fmt.Println(c, "open")
		}
	})

	go func() {
		// This one asks the connection for a connection instance
		// This will block if reconnection is in progress
		// This one guarantees that connection was alive just before the func is called
		c.WithOpenConnection(func(c *robbit.Connection) {
			fmt.Println("I'm given a connection", c.Id)
		})

		c.WithOpenChannel("source", func(c *robbit.Connection, channel *robbit.Channel) {
			fmt.Println("I'm given a channel", channel.Key)
		})
	}()

	c.RunForever()
}
