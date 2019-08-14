package topology

type Exchange struct {
	Name string
	Kind string
	Durable bool
	AutoDelete bool
	Internal bool
	NoWait bool
	Args map[string]interface{}
}

func (t *Exchange) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type bufferType Exchange
	b := &bufferType{}

	b.Kind = "fanout"
	b.Durable = true
	b.AutoDelete = false
	b.Internal = false
	b.NoWait = false
	b.Args = nil

	err := unmarshal(b)

	if b.Name == "" {
		panic("Empty name isn't allowed for exchange!")
	}
	if b.Kind != "fanout" && b.Kind != "direct" && b.Kind != "topic" && b.Kind != "headers" {
		panic("Exchange kind " + b.Kind + " is not recognized")
	}

	*t = Exchange(*b)
	return err
}
