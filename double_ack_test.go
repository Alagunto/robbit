package robbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"testing"
)

func TestConnection_Failure(t *testing.T) {
	ConnectTo("amqp://localhost:5672/").MaintainChannel("source", func(channel *amqp.Channel, connection *amqp.Connection) {
		_, err := channel.QueueDeclare("bug_test", true, true, false, false, map[string]interface{}{})
		if err != nil {
			panic(err)
		}

		err = channel.QueueBind("bug_test", "", "amq.fanout", false, map[string]interface{}{})
		if err != nil {
			panic(err)
		}

		c, err := channel.Consume("bug_test", "", false, true, true, false, map[string]interface{}{})
		if err != nil {
			panic(err)
		}

		not := make(chan *amqp.Error)
		connection.NotifyClose(not)

		go func() {
			for msg := range c {
				println(string(msg.Body))

				err := channel.Ack(msg.DeliveryTag, false)
				fmt.Printf("%v\n", err)
				fmt.Println("Acked")
				err = channel.Ack(msg.DeliveryTag, false)
				fmt.Printf("%v\n", err)
			}

			err = connection.Close()
			fmt.Println("Closed connection")
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		}()
	}).RunForever()
}
