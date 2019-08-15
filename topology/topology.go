package topology

type Topology struct {
	Exchanges []Exchange
	Queues    []Queue
	Bindings  []Binding
	Channels  []Channel

	ChannelForDeclarations Channel
}

func (t *Topology) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type bufferType Topology
	b := &bufferType{}

	b.ChannelForDeclarations = "service"

	err := unmarshal(b)

	serviceChannelExists := false
	for _, val := range b.Channels {
		if val == b.ChannelForDeclarations {
			serviceChannelExists = true
		}
	}

	if !serviceChannelExists {
		b.Channels = append(b.Channels, b.ChannelForDeclarations)
	}

	*t = Topology(*b)
	return err
}
