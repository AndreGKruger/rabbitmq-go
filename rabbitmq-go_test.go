package rabbitmqgo

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func setupRabbitMQ() RabbitMQ {
	config := RabbitMQConfig{
		ctx:            context.Background(),
		debug:          false,
		durable:        false,
		connection_url: "amqp://guest:guest@localhost:5672/",
		queue_name:     "test_queue",
		producer_tps:   100,
		consumer_tps:   100,
		config:         amqp.Config{},
	}
	return NewRabbitMq(config)
}

func TestPublish(t *testing.T) {
	r := setupRabbitMQ()

	// Publish a message
	msg := []byte(`{"message":"hello world"}`)
	err := r.Publish(msg)
	if err != nil {
		t.Fatalf("Error publishing message: %v", err)
	}
}
