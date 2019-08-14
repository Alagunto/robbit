package robbit

import (
	"fmt"
	"git-02.t1-group.ru/go-modules/utils"
	"github.com/streadway/amqp"
	"sync"
)

type Connection struct {
	RmqUrl       string

	connection *amqp.Connection

	channelsRequested  map[string]func(channel *amqp.Channel)
	callbacksRequested []func(*amqp.Connection, map[string]*amqp.Channel)

	openChannels map[string]*amqp.Channel
	connectionEstablished chan struct{}

	waitingGuys []func(connection *amqp.Connection)
	waitingGuysLock sync.Mutex

	connectionPrepared bool
}

// Returns a new connection to RabbitMQ, uninitialized yet
func ConnectTo(RmqUrl string) (connection *Connection) {
	connection = &Connection{
		RmqUrl: RmqUrl,
	}

	connection.channelsRequested = map[string]func(channel *amqp.Channel){}

	connection.openChannels = map[string]*amqp.Channel{}
	connection.connectionEstablished = make(chan struct{}, 128)

	return
}

func (c *Connection) WithOpenConnection(callback func(c *amqp.Connection)) {
	c.awaitPreparedConnection(func(connection *amqp.Connection) {
		callback(connection)
	})
}

func (c *Connection) WithOpenChannel(channelName string, callback func(c *amqp.Channel)) {
	c.awaitPreparedConnection(func(connection *amqp.Connection) {
		callback(c.openChannels[channelName])
	})
}

func (c *Connection) MaintainChannel(channelName string, initializingCallback func(channel *amqp.Channel)) *Connection {
	c.channelsRequested[channelName] = initializingCallback
	return c
}

func (c *Connection) InitializeWith(callback func(connection *amqp.Connection, channels map[string]*amqp.Channel)) *Connection {
	c.callbacksRequested = append(c.callbacksRequested, callback)

	return c
}

func (c *Connection) RunForever() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered after panic in main rmq loop")
			c.connection = nil
			c.RunForever() // Forfeiting our run loop, recovering connection via recursion, leaking miserably small window size
		}
	}()

	c.loop()
}

func (c *Connection) Run() *Connection {
	go c.RunForever()

	return c
}

func (c *Connection) loop() {
	// In case we panic in for-loop

	for {
		c.connectionPrepared = false // It's definitely not prepared yet

		fmt.Println("Knocking at " + c.RmqUrl + "...")
		connection, err := amqp.Dial(c.RmqUrl)
		utils.FailOnError(err, "Could not connect to RMQ at URL " + c.RmqUrl)
		if err != nil { // Retry
			continue
		}

		c.connection = connection
		c.openChannels = map[string]*amqp.Channel{}

		notify := connection.NotifyClose(make(chan *amqp.Error))

		weGood := true
		group := sync.WaitGroup{}
		group.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					weGood = false
					group.Done()
				}
			}()

			for channel, callback := range c.channelsRequested {
				callback(c.makeChannel(channel))
			}

			for _, callback := range c.callbacksRequested {
				callback(c.connection, c.openChannels)
			}

			group.Done()
		}()

		group.Wait()
		if weGood {
			c.connectionPrepared = true

			c.waitingGuysLock.Lock()
			defer c.waitingGuysLock.Unlock()

			fmt.Println("We're ok, notifying all waiting guys...")
			for _, guy := range c.waitingGuys {
				go guy(c.connection)
			}

			c.waitingGuys = []func(connection *amqp.Connection){}
		} else {
			// Someone panicked in the initialization
			// We shall just continue and reinit all the things
			println("Panic occured in the initialization of connection, recreating connection...")
			continue
		}

	ReceiveLoop:
		for {
			select {
			case err = <-notify:
				utils.FailOnError(err, "RMQ disconnected, reconnecting")
				c.connection = nil // Destroy it!
				break ReceiveLoop
			}
		}
	}
}


func (c *Connection) awaitPreparedConnection(callback func(connection *amqp.Connection)) {
	if c.connectionPrepared && (c.connection != nil) && !c.connection.IsClosed() {
		// We are ok most times
		callback(c.connection)
	} else {
		// We must wait for connection
		fmt.Println("Woopsie, I have to wait for the connection to get established")
		c.waitingGuysLock.Lock()
		defer c.waitingGuysLock.Unlock()

		c.waitingGuys = append(c.waitingGuys, callback)
	}
}

func (c *Connection) makeChannel(channelName string) *amqp.Channel {
	newChannel, err := c.connection.Channel()
	utils.FailOnError(err, "Cannot open channel on a given connection, woopsie")
	if err != nil {
		return nil
	}

	c.openChannels[channelName] = newChannel

	return newChannel
}
