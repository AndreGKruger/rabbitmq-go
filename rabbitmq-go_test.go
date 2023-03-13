package rabbitmqgo

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func setupRabbitMQ() RabbitMQ {
	config := RabbitMQConfig{
		Ctx:            context.Background(),
		Debug:          false,
		Durable:        false,
		Connection_url: "amqp://guest:guest@localhost:5672/",
		Queue_name:     "test_queue",
		Producer_tps:   100,
		Consumer_tps:   100,
		Config:         amqp.Config{},
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
