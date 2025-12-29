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

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handleGameLog,
	)
	if err != nil {
		log.Fatalf("could not subscribe to %v: %v", routing.GameLogSlug, err)
	}
	fmt.Printf("Subscribed to %v\n", routing.GameLogSlug)

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	gamelogic.PrintServerHelp()
game_loop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
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

func handleGameLog(gamelog routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(gamelog)
	if err != nil {
		log.Printf("Error writing log: %v", err)
		return pubsub.NackDiscard
	}
	return pubsub.Ack
}
