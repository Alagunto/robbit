package robbit

import (
	"github.com/alagunto/robbit/topology"
	"gopkg.in/yaml.v2"
)

func ApplyTopology(topology *topology.Topology, connector *Connector) {
	connector.MaintainChannel(string(topology.ChannelForDeclarations), func(connection *Connection, channel *Channel) {
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

		for _, binding := range topology.Bindings {
			err := channel.QueueBind(binding.QueueName, binding.Key, binding.Exchange, binding.NoWait, binding.Args)

			if err != nil {
				panic(err)
			}
		}
	})

	for _, channel := range topology.Channels {
		connector.MaintainChannel(string(channel), func(connection *Connection, channel *Channel) {})
	}
}

func DecomposeTopology(t string) (result *topology.Topology) {
	if t == "" {
		panic("Topology contents is empty. Probably, file is not found.")
	}

	result = &topology.Topology{}

	err := yaml.Unmarshal([]byte(t), result)

	if err != nil {
		panic(err)
	}

	return
}

