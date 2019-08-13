package robbit

import (
	"github.com/streadway/amqp"
	"testing"
	"time"
)

func TestConnectTo(t *testing.T) {
	c := ConnectTo("amqp://localhost:5672/").MaintainChannel("source", func(channel *amqp.Channel) {
		println("Channel ", channel, " is given ")
	})


	go func() {
		c.WithOpenConnection(func(c *amqp.Connection) {
			println("I've got a connection ", c)
		})

		time.Sleep(5 * time.Second)

		c.WithOpenConnection(func(c *amqp.Connection) {
			println("I've got a connection ", c)
		})
	}()

	c.RunForever()
}