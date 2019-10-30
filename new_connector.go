package robbit

import (
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"
)

type Connector struct {
	Credentials

	CurrentConnection *Connection

	Reconnect struct {
		*sync.Mutex
	}

	Callbacks struct {
		initialization  []Callback
		maintainChannel map[string]CallbackWithChannel

		*sync.Mutex
	}

	loopChannels struct {
		reconnectRequests   chan struct{}
		healthcheckRequests chan struct{}
	}

	connectionAwaiting struct {
		Queue []chan struct{}

		*sync.Mutex
	}
}

// Setting up callbacks for connection open
func (c *Connector) MaintainChannel(channel string, callback CallbackWithChannel) *Connector {
	c.Callbacks.maintainChannel[channel] = callback
	return c
}

func (c *Connector) InitializeWith(callback Callback) *Connector {
	c.Callbacks.initialization = append(c.Callbacks.initialization, callback)
	return c
}

func (c *Connector) WithOpenConnection(callback Callback) {
	c.awaitConnection(func(connection *Connection) {
		callback(connection)
	})
}

func (c *Connector) WithOpenChannel(channel string, callback CallbackWithChannel) {
	c.awaitConnection(func(connection *Connection) {
		callback(connection, connection.OpenChannels[channel])
	})
}

// Initialize the connection, open all desired channels on it, call all the callbacks
func (c *Connector) initializeConnection(connection *Connection) *Connection {
	c.Callbacks.Lock()
	defer c.Callbacks.Unlock()

	// Prepare channel
	for channelName, callback := range c.Callbacks.maintainChannel {
		channel := connection.OpenChannel(channelName)
		println("Registering channel ", channelName)
		callback(connection, channel)
	}

	connection.OpenChannel("healthcheck")

	for _, callback := range c.Callbacks.initialization {
		callback(connection)
	}

	connection.EnableNotificationChannels()
	connection.Prepared = true

	return connection
}

func (c *Connector) currentConnection() *Connection {
	return c.CurrentConnection
}

func (c *Connector) reconnect() {
	defer func() {
		r := recover()

		if r != nil {
			fmt.Println("Couldn't reconnect =(")
			c.requestReconnection()
		}
	}()

	c.Reconnect.Lock()
	defer c.Reconnect.Unlock()

	// Purge connection
	if c.CurrentConnection != nil {
		c.CurrentConnection.Purge()
	}

	// Initialize it, let everyone know we're ok
	c.CurrentConnection = NewConnection(c.Credentials.RmqConnection)

	c.initializeConnection(c.CurrentConnection)

	c.awakeWaitingQueue()

	c.loopChannels.reconnectRequests = make(chan struct{}, 128)
}

func (c *Connector) healthcheck() bool {
	if !c.CurrentConnection.IsReady() {
		// Don't healthcheck closed connections, mkay?
		return true
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	good := false

	c.WithOpenChannel("healthcheck", func(connection *Connection, channel *Channel) {
		defer func() {
			r := recover()

			if r != nil {
				fmt.Printf("Recovered from %v", r)
			}
		}()

		err := channel.Qos(10, 0, false)

		if err == nil {
			good = true
		} else {
			panic(err)
		}

		wg.Done()
	})

	fmt.Println("Healthcheck", good)

	return good
}

func (c *Connector) awakeWaitingQueue() {
	c.connectionAwaiting.Lock()
	defer c.connectionAwaiting.Unlock()

	for _, c := range c.connectionAwaiting.Queue {
		c <- struct{}{}
	}

	c.connectionAwaiting.Queue = []chan struct{}{}
}

func (c *Connector) awaitConnection(callback func(*Connection)) {
	// We want to reconnect if our callback caused troubles
	defer func() {
		recovered := recover()
		if recovered != nil {
			fmt.Println("Recovered from some kind of panic")
			fmt.Printf("%v", recovered)
			c.requestReconnection()
		}
	}()

	// We don't really want to cross with the reconnect
	c.Reconnect.Lock()
	defer c.Reconnect.Unlock()

	// If we are connected, life is really simple
	if (c.CurrentConnection != nil) && (c.CurrentConnection.IsReady()) {
		callback(c.CurrentConnection)
	} else {
		fmt.Println("We are not connected, awaiting")
		// We are creating a channel and asking the reconnection system to notify us when we're good
		go func() {
			defer func() {
				// We want to reconnect if our callback caused troubles
				recovered := recover()
				if recovered != nil {
					fmt.Println("Recovered from some kind of panic while awaiting for the connection")
					fmt.Printf("%v", recovered)
					c.requestReconnection()
				}
			}()

			c.connectionAwaiting.Lock()
			defer c.connectionAwaiting.Unlock()

			ch := make(chan struct{})
			c.connectionAwaiting.Queue = append(c.connectionAwaiting.Queue, ch)

			fmt.Println("Locked till we good")
			<-ch // Lock until we are good

			callback(c.CurrentConnection)
		}()
	}
}

func (c *Connector) requestReconnection() {
	c.loopChannels.reconnectRequests <- struct{}{}
}

func (c *Connector) cycle() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println("Panicked in robbit cycle")
			fmt.Printf("%v\n", err)
			c.requestReconnection()

			panic(err)
		}
	}()

	select {
		case _ = <-c.loopChannels.healthcheckRequests: {
			fmt.Println("Healthcheck going on")
			go func() {
				ok := c.healthcheck()
				if !ok {
					c.requestReconnection()
				}
			}()
		}

		case _ = <-c.loopChannels.reconnectRequests: {
			fmt.Println("RMQ reconnects due to manual reconnect request")
			c.reconnect()
		}
		default:
	}

	connection := c.currentConnection()
	if connection != nil {
		select {
			case _ = <-connection.ErrorNotifications: {
				fmt.Println("RMQ reconnects due to error notification from amqp")
				c.requestReconnection()
			}
			case _ = <-c.currentConnection().BlockingNotification: {
				fmt.Println("Blocking notification from amqp")
			}
			default:
		}
	}
}

func (c *Connector) WithTopologyFrom(reader io.Reader) *Connector {
	s, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	ApplyTopology(DecomposeTopology(string(s)), c)

	return c
}

func (c *Connector) EnableHealthchecks() {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Tick at", t)
				c.loopChannels.healthcheckRequests <- struct{}{}
			}
		}
	}()
}

func (c *Connector) RunForever() {
	fmt.Println("Running forever")
	c.requestReconnection()

	for {
		c.cycle()
	}
}

func To(connectionString string) *Connector {
	connector := &Connector{
		loopChannels: struct {
			reconnectRequests   chan struct{}
			healthcheckRequests chan struct{}
		}{reconnectRequests: make(chan struct{}, 128), healthcheckRequests: make(chan struct{}, 128)},
		Reconnect:         struct{ *sync.Mutex }{Mutex: &sync.Mutex{}},
		CurrentConnection: nil,
		Callbacks: struct {
			initialization  []Callback
			maintainChannel map[string]CallbackWithChannel
			*sync.Mutex
		}{initialization: []Callback{}, maintainChannel: map[string]CallbackWithChannel{}, Mutex: &sync.Mutex{}},
		connectionAwaiting: struct {
			Queue []chan struct{}
			*sync.Mutex
		}{Queue: []chan struct{}{}, Mutex: &sync.Mutex{}},
		Credentials: Credentials{RmqConnection: connectionString},
	}

	connector.EnableHealthchecks()
	return connector
}
