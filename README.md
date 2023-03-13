# RabbitMQGo Package
A package to provide a simplified interface to publish and consume messages from RabbitMQ, using the amqp091-go library by RabbitMQ.

## Installation
To use the RabbitMQGo package in your Go project, simply run the following command:

``` go get -u github.com/AndreGKruger/rabbitmq-go ```

## To use this package, import it as shown below:

```
import (
    "github.com/AndreGKruger/rabbitmqgo"
)
```

## Initialization
The rabbitmqgo.NewRabbitMq() function is used to initialize a new instance of the RabbitMQ struct. The required configuration parameters are set through the RabbitMQConfig{} struct.

Example:

```
config := rabbitmqgo.RabbitMQConfig{
    ctx:            context.Background(),
    debug:          true,
    durable:        true,
    connection_url: "amqp://guest:guest@localhost:5672/",
    queue_name:     "test",
    producer_tps:   100,
    consumer_tps:   100,
    config:         amqp.Config{},
}

rabbitmqInstance := rabbitmqgo.NewRabbitMq(config)
```

## Publishing Messages
Use the Publish([]byte) error method to send messages to the designated queue.

Example:

```
message := []byte(`{"data": "This is a message"}`)
err := rabbitmqInstance.Publish(message)

if err != nil {
    log.Fatal(err)
}
```

## Consuming Messages
Use the Consume(string,int,func([]byte) error)error method to retrieve messages from a queue. The method accepts three parameters:

queue name (string)
prefetch count (int)
handler function for processing messages (func([]byte) error)
Example:

```
const autoAck = false
const concurrentHandlers = 1
const prefetchCount = 5

handler := func(body []byte) error {
    log.Println("Message received: ", string(body))
    return nil
}

err = rabbitmqInstance.Consume("consumer", prefetchCount,
    rabbitmqgo.HandlerWrapper(concurrentHandlers,handler,autoAck),
)

if err != nil {
    log.Fatal(err)
}
```

## Retrieving Queue Statistics
Use the GetStats() (amqp.Queue, error) method to fetch statistics (amqp.Queue struct) for the specified queue.
Example:

```
stats, err := rabbitmqInstance.GetStats()
if err != nil {
    log.Fatal(err)
}

log.Println("Stats:\nName :", stats.Name,
            "\nMessages :", stats.Messages,
            "\nConsumers :", stats.Consumers))
```

Note that detailed documentation and working examples can be found in the godoc for this library.