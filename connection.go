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

	channelWrapper := &Channel{Channel: *channel}

	c.OpenChannels[name] = channelWrapper

	return channelWrapper
}

func (c *Connection) EnableNotificationChannels() {
	c.ErrorNotifications = make(chan *amqp.Error, 128)
	c.BlockingNotification = make(chan amqp.Blocking, 128)
	c.NotifyClose(c.ErrorNotifications)
	c.NotifyBlocked(c.BlockingNotification)
	println("Notification would be sent to my channel lol")

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

	println("Purging stuff")
	//for _, channel := range c.OpenChannels {
	//	_ = channel.Close() // Errors might be
	//}

	if !c.IsClosed() {
		_ = c.Close()
	}
	println("Purged")
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
