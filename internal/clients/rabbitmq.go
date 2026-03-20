package clients

import (
	"context"
	"fmt"
	"transfers-api/internal/config"
	"transfers-api/internal/logging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn  *amqp.Connection
	queue string
}

func NewRabbitMQClient(cfg config.RabbitMQ) *RabbitMQClient {
	conn, err := amqp.Dial(
		fmt.Sprintf(
			"amqp://%s:%s@%s:%d/",
			cfg.Username,
			cfg.Password,
			cfg.Hostname,
			cfg.Port,
		),
	)
	if err != nil {
		logging.Logger.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	return &RabbitMQClient{
		conn:  conn,
		queue: cfg.QueueName,
	}
}

func (c *RabbitMQClient) Publish(operation string, transferID string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	body := fmt.Sprintf("%s:%s", operation, transferID)
	err = ch.Publish(
		"",      // exchange
		c.queue, // routing key (queue name)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (c *RabbitMQClient) Consume(ctx context.Context, handler func(ctx context.Context, body []byte) error) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	if err := c.ensureQueue(); err != nil {
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to configure qos: %w", err)
	}

	msgs, err := ch.Consume(
		c.queue,
		"transfers-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("rabbitmq deliveries channel closed")
			}

			if err := handler(ctx, msg.Body); err != nil {
				logging.Logger.Warnf("error processing RabbitMQ message: %v", err)
				if nackErr := msg.Nack(false, true); nackErr != nil {
					logging.Logger.Warnf("error sending nack: %v", nackErr)
				}
				continue
			}

			if err := msg.Ack(false); err != nil {
				logging.Logger.Warnf("error sending ack: %v", err)
			}
		}
	}
}

func (c *RabbitMQClient) ensureQueue() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	return nil
}

func (c *RabbitMQClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}
