package pubsub

import (
	"encoding/json"

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
