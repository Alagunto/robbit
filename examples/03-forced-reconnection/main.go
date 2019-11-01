package main

import (
	"git-02.t1-group.ru/go-modules/robbit"
)

/*
	This one shows ways to force a reconnection
*/
func main() {
	c := robbit.To("amqp://localhost:5672/")

	c.MaintainChannel("source", func(connection *robbit.Connection, channel *robbit.Channel) {
		//fmt.Println("Maintaining the channel...")
		_ = connection.Close() // This will make robbit die
	})

	c.RunForever()

}
