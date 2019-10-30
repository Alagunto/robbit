package robbit

type Callback func(connection *Connection)

type CallbackWithChannel func(connection *Connection, channel *Channel)