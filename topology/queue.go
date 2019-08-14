package topology

type Queue struct {
	Name string
	Durable bool
	AutoDelete bool
	Exclusive bool
	NoWait bool
	Args map[string]interface{}
}

func (t *Queue) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type bufferType Queue
	b := &bufferType{}

	b.Durable = true
	b.AutoDelete = false
	b.Exclusive = false
	b.NoWait = false
	b.Args = nil

	err := unmarshal(b)

	if b.Name == "" {
		panic("Queue name cannot be empty!")
	}

	*t = Queue(*b)
	return err
}
