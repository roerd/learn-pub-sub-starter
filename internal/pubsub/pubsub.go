package pubsub

import (
	"bytes"
	"encoding/gob"
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
	queueArgs := amqp.Table{}
	queueArgs["x-dead-letter-exchange"] = "peril_dlx"
	q, err := ch.QueueDeclare(queueName, simpleQueueType == Durable, simpleQueueType == Transient, simpleQueueType == Transient, false, queueArgs)
	if err != nil {
		return ch, q, err
	}
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return ch, q, err
	}
	return ch, q, err
}

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
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
			val, err := unmarshaller(msg.Body)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Ack(false)
				continue
			}
			acktype := handler(val)
			switch acktype {
			case Ack:
				msg.Ack(false)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()
	return nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	unmarshaller := func(body []byte) (T, error) {
		var val T
		err := json.Unmarshal(body, &val)
		return val, err
	}
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaller)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	unmarshaller := func(body []byte) (T, error) {
		var val T
		buf := bytes.NewBuffer(body)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&val)
		return val, err
	}
	return subscribe(conn, exchange, queueName, key, simpleQueueType, handler, unmarshaller)
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(val)
	if err != nil {
		return err
	}
	val_gob := buf.Bytes()

	return ch.Publish(exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        val_gob,
	})
}
