package main

import (
	"fmt"
	"git-02.t1-group.ru/go-modules/robbit"
	"github.com/streadway/amqp"
)

/*
	This one shows ways to force a reconnection
*/
func main() {
	c := robbit.ConnectTo("amqp://localhost:5672/")

	c.MaintainChannel("source", func(channel *amqp.Channel, connection *amqp.Connection) {
		//fmt.Println("Maintaining the channel...")
		_ = connection.Close() // This will make robbit recreate everything!
	})

	c.InitializeWith(func(connection *amqp.Connection, channels map[string]*amqp.Channel) {
		fmt.Println("Initializing...")
		// or you can just...
		c.Reconnect()
	})

	c.RunForever()

}
