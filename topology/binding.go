package topology

type Binding struct {
	QueueName string
	Key string
	Exchange string
	NoWait bool
	Args map[string]interface{}
}

func (t *Binding) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type bufferType Binding
	b := &bufferType{}

	b.Key =  ""
	b.NoWait = false
	b.Args = nil

	err := unmarshal(b)

	if b.Exchange == "" {
		panic("Binding exchange cannot be empty!")
	}

	if b.QueueName == "" {
		panic("Binding queue name cannot be empty!")
	}

	*t = Binding(*b)
	return err
}