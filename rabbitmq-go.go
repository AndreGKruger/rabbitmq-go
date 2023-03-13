package rabbitmqgo

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/ratelimit"
)

type RabbitMQ interface {
	Publish(message []byte) error
	Consume(name string, prefetch int, handler func([]byte) error) error
	GetStats() (amqp.Queue, error)
	initChannel() error
	debuglog(message string, args ...interface{})
}

type rabbitmq struct {
	ctx            context.Context
	logger         *log.Logger
	debug          bool
	durable        bool
	connection_url string
	config         amqp.Config
	connection     *amqp.Connection
	channel        *amqp.Channel
	queue_name     string
	producer_tps   ratelimit.Limiter
	consumer_tps   ratelimit.Limiter
}

type RabbitMQConfig struct {
	Ctx            context.Context
	Debug          bool
	Durable        bool
	Connection_url string
	Queue_name     string
	Producer_tps   int
	Consumer_tps   int
	Config         amqp.Config
}

func NewRabbitMq(cnf RabbitMQConfig) RabbitMQ {
	return &rabbitmq{
		ctx:            cnf.Ctx,
		logger:         log.Default(),
		debug:          cnf.Debug,
		durable:        cnf.Durable,
		connection_url: cnf.Connection_url,
		config:         cnf.Config,
		queue_name:     cnf.Queue_name,
		producer_tps:   ratelimit.New(cnf.Producer_tps),
		consumer_tps:   ratelimit.New(cnf.Consumer_tps),
	}
}

func (r *rabbitmq) Publish(message []byte) error {
	// Check if the channel is initialized
	if r.channel == nil {
		r.debuglog("Channel is not initialized, initializing now")
		err := r.initChannel()
		if err != nil {
			return err
		}
	}

	prev := time.Now()
	now := r.producer_tps.Take()
	err := r.channel.PublishWithContext(
		r.ctx,
		"",
		r.queue_name,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/json",
			ContentEncoding: "",
			Body:            message,
			DeliveryMode:    amqp.Persistent,
			Priority:        0,
		},
	)
	now.Sub(prev)
	prev = now
	r.debuglog("Message published to RabbitMQ")
	return err
}

func (r *rabbitmq) Consume(name string, prefetch int, handler func([]byte) error) error {
	// Check if the channel is initialized
	if r.channel == nil {
		r.debuglog("Channel is not initialized, initializing now")
		err := r.initChannel()
		if err != nil {
			return err
		}
	}

	// Set Channel Qos prefetch count
	err := r.channel.Qos(prefetch, 0, false)
	if err != nil {
		r.debuglog("[RabbitMq] - Error setting channel Qos")
		return err
	}

	// Consume messages
	r.debuglog("[RabbitMq] - Consuming messages")
	done := make(chan error)
	go func() {
		msgs, err := r.channel.Consume(r.queue_name, name, false, false, false, false, nil)
		if err != nil {
			r.debuglog("[RabbitMq] - Error consuming messages: %s", err.Error())
			done <- err
		}

		prev := time.Now()
		for d := range msgs {
			now := r.consumer_tps.Take()
			err := handler(d.Body)
			if err != nil {
				r.debuglog("[RabbitMq] - Error handling message: %s", err.Error())
				r.debuglog("[RabbitMq] - Rejecting message")
				d.Reject(false)
				now.Sub(prev)
				prev = now
				done <- err
			}
			d.Ack(false)
			now.Sub(prev)
			prev = now
		}
	}()
	return <-done
}

func (r *rabbitmq) GetStats() (amqp.Queue, error) {
	// Check if the channel is initialized
	if r.channel == nil {
		r.debuglog("Channel is not initialized, initializing now")
		err := r.initChannel()
		if err != nil {
			return amqp.Queue{}, err
		}
	}

	queue, err := r.channel.QueueInspect(r.queue_name)
	if err != nil {
		r.debuglog("Error getting queue info: %s", err.Error())
		return amqp.Queue{}, err
	}

	return queue, nil
}

// PRIVATE METHODS ---------------------------------------------------------------------------------------------
func (r *rabbitmq) initChannel() error {
	var err error

	// Open a new connection if it hasn't been opened yet
	r.debuglog("Opening a new connection to RabbitMQ")
	if r.connection == nil || r.connection.IsClosed() {
		r.connection, err = amqp.DialConfig(r.connection_url, r.config)
		if err != nil {
			return err
		}
	}

	// Open a new channel
	r.debuglog("Opening a new channel to RabbitMQ")
	r.channel, err = r.connection.Channel()
	if err != nil {
		return err
	}

	// Declare the queue
	_, err = r.channel.QueueDeclare(
		r.queue_name,
		r.durable, // durable
		false,     // autoDelete
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		r.debuglog("Error declaring queue: %s", err.Error())
		return err
	}
	r.debuglog("Queue declared")
	return nil
}

// private function to write logs when debug is true taking in message and args
func (r *rabbitmq) debuglog(message string, args ...interface{}) {
	if !r.debug {
		return
	}
	r.logger.Print(message, args)
}
