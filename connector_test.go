package robbit

import (
	"github.com/streadway/amqp"
	"os"
	"testing"
)

func TestConnectTo(t *testing.T) {
	c := ConnectTo("amqp://localhost:5672/").MaintainChannel("source", func(channel *amqp.Channel) {
		println("Channel ", channel, " is given ")
	})

	go func() {
		c.WithOpenConnection(func(c *amqp.Connection) {
			println("I've got a connection ", c)
		})
	}()

	c.RunForever()
}

func TestConnection_WithTopologyFrom(t *testing.T) {
	f, err := os.Open("./topology/test-examples/00-simple.yaml")

	if err != nil {
		panic(err)
	}

	c := ConnectTo("amqp://localhost:5672/").WithTopologyFrom(f).InitializeWith(func(connection *amqp.Connection, channels map[string]*amqp.Channel) {
		if _, exists := channels["one"]; !exists {
			t.Error("No channel 'one'")
		}

		if _, exists := channels["two"]; !exists {
			t.Error("No channel 'two'")
		}

		if _, exists := channels["lol"]; !exists {
			t.Error("No channel 'lol'")
		}
	})

	c.RunForever()
}