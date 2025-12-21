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
	fmt.Println("Starting Peril client...")

	const url = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("Connection was successful")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gamestate := gamelogic.NewGameState(username)

	pauseQueue := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, pauseQueue, routing.PauseKey, pubsub.Transient)
	pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, pauseQueue, routing.PauseKey, pubsub.Transient, handlerPause(gamestate))

	moveQueue := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)
	moveKey := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, "*")
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, moveQueue, moveKey, pubsub.Transient, handlerMove(gamestate))

game_loop:
	for {
		words := gamelogic.GetInput()
		switch words[0] {
		case "spawn":
			err := gamestate.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			move, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Moving worked")
			}
			ch, err := conn.Channel()
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, moveQueue, move)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Publishing move worked")
			}
			ch.Close()
		case "status":
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break game_loop
		default:
			fmt.Println("Command not understood")
		}
	}
	fmt.Println("Shutting down...")
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		gs.HandleMove(move)
	}
}
