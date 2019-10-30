package robbit

import (
	//"github.com/streadway/amqp"
	"os"
	"testing"
	"time"
)

func TestNewConnector(t *testing.T) {
	c := To("amqp://localhost:5672/").MaintainChannel("source", func(connection *Connection, channel *Channel) {})
	go func() {
		c.RunForever()
	}()

	ch := time.Tick(time.Second * 2)

	for _ = range ch {
		c.WithOpenConnection(func(c *Connection) {
			//err := c.OpenChannels["source"].Publish("amq.fanout", "", true, false, struct {
			//	Headers         amqp.Table
			//	ContentType     string
			//	ContentEncoding string
			//	DeliveryMode    uint8
			//	Priority        uint8
			//	CorrelationId   string
			//	ReplyTo         string
			//	Expiration      string
			//	MessageId       string
			//	Timestamp       time.Time
			//	Type            string
			//	UserId          string
			//	AppId           string
			//	Body            []byte
			//}{Headers: map[string]interface{}{}, Body: []byte("kek")})

			//if err != nil {
			//	panic(err)
			//}
		})
	}

}

func TestConnector_TopologyFrom(t *testing.T) {
	f, err := os.Open("./topology/test-examples/00-simple.yaml")

	if err != nil {
		panic(err)
	}

	c := To("amqp://localhost:5672/").WithTopologyFrom(f).InitializeWith(func(connection *Connection) {
		if _, exists := connection.OpenChannels["one"]; !exists {
			t.Error("No channel 'one'")
		}

		if _, exists := connection.OpenChannels["two"]; !exists {
			t.Error("No channel 'two'")
		}

		if _, exists := connection.OpenChannels["lol"]; !exists {
			t.Error("No channel 'lol'")
		}

	})

	c.RunForever()
}