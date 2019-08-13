package main

import "github.com/streadway/amqp"
import "git-02.t1-group.ru/go-modules/robbit"

func main() {
	// It declares a new connection, not actually connecting yet
	c := robbit.ConnectTo("amqp://localhost:5672/")

	// It tells connection to create a channel each time it's reconnecting.
	// After each channel creation, given callback will be run
	c.MaintainChannel("source", func(channel *amqp.Channel) {
		println("Channel", channel, "is given")
	})

	go func() {
		// This one asks the connection for a connection instance
		// This will block if reconnection is in progress
		// This one guarantees that connection was alive just before the func is called
		c.WithOpenConnection(func(c *amqp.Connection) {
			println("I'm given a connection", c)
		})

		c.WithOpenChannel("source", func(c *amqp.Channel) {
			println("I'm given a channel", c)
		})
	}()

	c.RunForever()
}
