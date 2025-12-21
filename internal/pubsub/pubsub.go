package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	val_json, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return ch.Publish(exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        val_json,
	})
}

type SimpleQueueType int

const (
	Durable SimpleQueueType = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return ch, amqp.Queue{}, err
	}
	q, err := ch.QueueDeclare(queueName, simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient, false, nil)
	if err != nil {
		return ch, q, err
	}
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return ch, q, err
	}
	return ch, q, err
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for msg := range msgs {
			var val T
			err := json.Unmarshal(msg.Body, &val)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Ack(false)
				continue
			}
			handler(val)
			msg.Ack(false)
		}
	}()
	return nil
}
