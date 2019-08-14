package topology

import (
	"git-02.t1-group.ru/go-modules/robbit"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
)

func Apply(topology *Topology, connection *robbit.Connection) {
	// First, exchanges
	connection.MaintainChannel(string(topology.ChannelForDeclarations), func(channel *amqp.Channel) {
		for _, exchange := range topology.Exchanges {
			err := channel.ExchangeDeclare(
				exchange.Name,
				exchange.Kind,
				exchange.Durable,
				exchange.AutoDelete,
				exchange.Internal,
				exchange.NoWait,
				exchange.Args,
			)

			if err != nil {
				panic(err)
			}
		}

		for _, queue := range topology.Queues {
			_, err := channel.QueueDeclare(
				queue.Name,
				queue.Durable,
				queue.AutoDelete,
				queue.Exclusive,
				queue.NoWait,
				queue.Args,
			)

			if err != nil {
				panic(err)
			}
		}

		//for _, binding := range topology.Binding {
		//	_, err := channel.QueueBind()
		//}
	})

	// Now, queues
}

func Decompose(topology string) (result *Topology) {
	result = &Topology{}

	err := yaml.Unmarshal([]byte(topology), result)

	if err != nil {
		panic(err)
	}

	return
}
