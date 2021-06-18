package robbit

import (
	"fmt"
	"github.com/streadway/amqp"
	"math/rand"
)

type Connection struct {
	amqp.Connection

	Id           int64
	OpenChannels map[string]*Channel

	Broken   bool
	Prepared bool

	ErrorNotifications   chan *amqp.Error
	BlockingNotification chan amqp.Blocking
}

func (c *Connection) OpenChannel(name string) *Channel {
	channel, err := c.Channel()
	if err != nil {
		panic(err)
	}

	channelWrapper := &Channel{Channel: *channel, Key: name}

	c.OpenChannels[name] = channelWrapper

	return channelWrapper
}

func (c *Connection) EnableNotificationChannels() {
	c.ErrorNotifications = make(chan *amqp.Error, 128)
	c.BlockingNotification = make(chan amqp.Blocking, 128)
	c.NotifyClose(c.ErrorNotifications)
	c.NotifyBlocked(c.BlockingNotification)
}

func (c *Connection) Purge() {
	defer func() {
		r := recover()
		if r != nil {
			fmt.Println("RMQ couldn't properly close the connection")
			fmt.Println(r)
		}
	}()

	c.Broken = true

	if !c.IsClosed() {
		_ = c.Close()
	}
}

func (c *Connection) IsReady() bool {
	return c.Prepared && !c.Connection.IsClosed() && !c.Broken
}

func NewConnection(connectionString string) *Connection {
	connection, err := amqp.Dial(connectionString)
	if err != nil {
		panic(err)
	}

	c := &Connection{
		OpenChannels: map[string]*Channel{},
		Id: rand.Int63(),
		Connection: *connection,
	}

	return c
}
