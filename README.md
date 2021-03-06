## Robbit, the RabbitMQ connector

Это очень маленький и простой модуль-обвязка вокруг amqp, который позволяет безболезненно подключаться к RabbitMQ и не бояться сбоев.

На данный момент библиотека, на которую опирается robbit, умеет падать без возможности восстановления.
К сожалению, на патчинг её нет времени, так что используя robbit будьте готовы к тому, что вся программа упадёт.

Зато не зависнет.

_он просто [почти] роби́т_
### Простой пример
```go
c := robbit.To("amqp://localhost:5672/")
c.RunForever() // blocking
```

`To(...)` создаёт объект подключения, но не осуществляет подключение, пока не будет вызван `Run` или `RunForever`. 
   
`RunForever()` запускает вечный цикл поддержания соединения. Если соединение прервётся, модуль сам перезапустит его.

### Каналы и переподключение
```go
c := robbit.To("amqp://localhost:5672/")
c.MaintainChannel("source", func(channel *robbit.Channel) {
    fmt.Println("Channel", channel, "is given")
})
```

`MaintainChannel` поддерживает объявленным именованный канал. Имя канала имеет смысл только в пределах этого модуля, оно не посылается в RabbitMQ.

Callback, который передаётся в `MaintainChannel`, вызывается каждый раз, когда соединение перезапускается, давая возможность выполнить какой-то код при открытии канала.

Из этого callback'а имеет смысл объявлять очереди, exchange'ы и бинды.

```go
c.InitializeWith(func(connection *robbit.Connection) {
    for c, _ := range connection.OpenChannels {
        fmt.Println(c, "open")
    }
})
```

`InitializeWith` позволяет установить callback, который будет вызван сразу после того, как все `MaintainChannel` callback'и были вызваны.

В этот callback следует добавлять логику, которая должна выполниться как только соединение будет открыто, но после того, как для этого соединения будут объявлены все сущности.

__Если любой из этих коллбеков запаникует, подключение будет перезапущено__ 

### Чтение 

```go
c := robbit.To("amqp://localhost:5672/")

c.MaintainChannel("source", func(channel *robbit.Channel) {}) 

c.InitializeWith(func(connection *robbit.Connection) {
    msgs, _ := channels["source"].Consume(
        "queue-name",        // queue
        "",                  // consumer
        false,               // auto-ack
        false,               // exclusive
        false,               // no-local
        false,               // no-wait
        nil,                 // args
    )
    
    go func() {
        for msg := range msgs {
            // ...
        }
    }()
})

c.RunForever()
```

### Запись

```go
c := robbit.To("amqp://localhost:5672/")

c.MaintainChannel("target", func(channel *robbit.Channel) {}) 

...

c.WithOpenChannel("target", func(c *robbit.Channel) {
    c.Publish("queue-name",
        "",    // routing key
        false, // mandatory
        false, // immediate
        amqp.Publishing{
            ContentType: "<...>", 
            Body: <...>,
        }
    )
})

c.RunForever()
```

`WithOpenChannel` позволяет получить актуальный объект канала. Гарантируется, что на момент вызова callback'а канал был открыт и доступен.
Если `WithOpenChannel` будет вызван во время переподключения, данный callback будет вызван только после того, как соединение установится, и выполнятся `MaintainChannel` и `InitializeWith` callback'и.

`WithOpenConnection` делает то же самое, но возвращает объект открытого соединения `*amqp.Connection`.

### Топология по конфиг-файлу

Чтобы не объявлять кучу очередей, биндов и прочих сущностей методом копипастинга `channel.DelcareBullshit`, можно сделать config-файл в yaml и подгрузить топологию из него.

__config.yaml:__
```yaml
exchanges:
  - name: fan
    kind: fanout

channels:
  - one
  - two

queues:
  - name: queue
    durable: true

bindings:
  - queuename: queue
    key: ""
    exchange: fan

channelfordeclarations: lol
```

Далее,

```go
topology, _ := os.Open("config.yaml")

c := robbit.To("amqp://localhost:5672/").
    WithTopologyFrom(topology) // ...и всё.
    		
c.RunForever()    		
```

Этот код создаст канал _channelfordeclarations_ через `MaintainChannel`, в нём объявит сначала exchange'ы, потом очереди, потом биндинги.

Все остальные каналы будут объявлены через `MaintainChannel`.

Объявлять несколько `MaintainChannel` на один канал можно — проблем не возникнет, канал будет создан лишь однажды, все callback'и получат один и тот же объект канала.

#### Важно

Если какие-то поля в конфиг-файле пропущены, будет использовано дефолтное значение, не обязательно пустое.

Дефолтные значения:

- Binding
    ```
    Key =  ""
    NoWait = false
    Args = nil
    ```
- Queue
    ```
    Durable = true
    AutoDelete = false
    Exclusive = false
    NoWait = false
    Args = nil
    ```
- Exchange
    ```
    Kind = "fanout"
    Durable = true
    AutoDelete = false
    Internal = false
    NoWait = false
    Args = nil
    ```
-  ChannelForDeclarations: `service`

#### Примеры

В папке examples есть пара примеров, можно посмотреть.

#### Тестирование

Модуль толком не  снабжён автоматическими тестами, а надо бы. Найдёте баг — feel free to fix или метните сообщение @alagunto в телегу.