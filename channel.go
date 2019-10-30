package robbit

import "github.com/streadway/amqp"

type Channel struct {
	amqp.Channel

	Key string
}
