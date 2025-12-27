package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	const url = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Connection was successful")

	_, queue, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
	)
	if err != nil {
		log.Fatalf("could not subscribe to %v: %v", routing.GameLogSlug, err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	gamelogic.PrintServerHelp()
game_loop:
	for {
		words := gamelogic.GetInput()
		switch words[0] {
		case "pause":
			fmt.Println("Sending pause message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
		case "resume":
			fmt.Println("Sending resume message")
			pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
		case "quit":
			fmt.Println("Shutting down...")
			break game_loop
		default:
			fmt.Println("Command not understood")
		}
	}
}
